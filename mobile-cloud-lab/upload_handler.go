package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func uploadHandler(ctx context.Context, event events.S3Event) (string, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Sprintf(`{"message":"Server error","error":"%v"}`, err), err
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})
	defer rdb.Close()

	var csvData string
	if os.Getenv("NODE_ENV") == "local" {
		data, err := os.ReadFile("products.csv")
		if err != nil {
			return fmt.Sprintf(`{"message":"Error reading local file","error":"%v"}`, err), err
		}
		csvData = string(data)
	} else {
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
		if err != nil {
			return fmt.Sprintf(`{"message":"AWS config error","error":"%v"}`, err), err
		}
		s3Client := s3.NewFromConfig(cfg)
		bucket := event.Records[0].S3.Bucket.Name
		key := event.Records[0].S3.Object.Key
		resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
		if err != nil {
			return fmt.Sprintf(`{"message":"S3 error","error":"%v"}`, err), err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Sprintf(`{"message":"Error reading S3 object","error":"%v"}`, err), err
		}
		csvData = string(body)
	}

	reader := csv.NewReader(strings.NewReader(csvData))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Sprintf(`{"message":"CSV parse error","error":"%v"}`, err), err
	}

	if len(records) <= 1 {
		return `{"message":"No valid products in CSV"}`, nil
	}

	products := []Product{}
	for i, row := range records[1:] {
		if len(row) < 5 {
			fmt.Printf("Skipping invalid row %d\n", i+1)
			continue
		}
		price, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Printf("Invalid price in row %d\n", i+1)
			continue
		}
		qty, err := strconv.Atoi(row[4])
		if err != nil {
			fmt.Printf("Invalid qty in row %d\n", i+1)
			continue
		}
		id := row[0]
		if id == "" {
			id = uuid.New().String()
		}
		var image *string
		if row[2] != "" {
			image = &row[2]
		}
		if row[1] == "" {
			continue
		}
		products = append(products, Product{ID: id, Name: row[1], Image: image, Price: price, Qty: qty})
	}

		if len(products) == 0 {
			return `{"message":"No valid products in CSV"}`, nil
		}

		for _, p := range products {
			_, err := db.ExecContext(ctx,
				`INSERT INTO products (id, name, image, price, qty, updated_at)
				VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
				ON CONFLICT (id) DO UPDATE
				SET name = EXCLUDED.name, image = EXCLUDED.image, price = EXCLUDED.price,
				qty = EXCLUDED.qty, updated_at = CURRENT_TIMESTAMP`,
				p.ID, p.Name, p.Image, p.Price, p.Qty)
			if err != nil {
				fmt.Printf("DB error for product %s: %v\n", p.ID, err)
				continue
			}
			data, err := json.Marshal(p)
			if err != nil {
				fmt.Printf("JSON marshal error for product %s: %v\n", p.ID, err)
				continue
			}
			err = rdb.Set(ctx, "product:"+p.ID, string(data), 0).Err()
			if err != nil {
				fmt.Printf("Redis error for product %s: %v\n", p.ID, err)
			}
		}

		return fmt.Sprintf(`{"message":"Products uploaded successfully","count":%d}`, len(products)), nil
}