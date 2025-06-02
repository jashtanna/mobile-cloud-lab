package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func getAllHandler(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})
	defer rdb.Close()

	products := []Product{}
	keys, err := rdb.Keys(ctx, "product:*").Result()
	if err == nil && len(keys) > 0 {
		for _, key := range keys {
			data, err := rdb.Get(ctx, key).Result()
			if err != nil {
				fmt.Printf("Redis get error for %s: %v\n", key, err)
				continue
			}
			var p Product
			if err := json.Unmarshal([]byte(data), &p); err != nil {
				fmt.Printf("Redis parse error for %s: %v\n", key, err)
				continue
			}
			products = append(products, p)
		}
	}

	if len(products) == 0 {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"message":"Server error","error":"%v"}`, err),
			}, err
		}
		defer db.Close()

		rows, err := db.QueryContext(ctx, "SELECT id, name, image, price, qty, out_of_stock, created_at, updated_at FROM products")
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"message":"DB error","error":"%v"}`, err),
			}, err
		}
		defer rows.Close()

		for rows.Next() {
			var p Product
			var image sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &image, &p.Price, &p.Qty, &p.OutOfStock, &p.CreatedAt, &p.UpdatedAt); err != nil {
				fmt.Printf("DB scan error: %v\n", err)
				continue
			}
			if image.Valid {
				p.Image = &image.String
			}
			products = append(products, p)
			data, _ := json.Marshal(p)
			rdb.Set(ctx, "product:"+p.ID, string(data), 0)
		}
	}

	body, err := json.Marshal(products)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"message":"JSON error","error":"%v"}`, err),
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Cache-Control": "public, max-age=300",
		},
		Body: string(body),
	}, nil
}
