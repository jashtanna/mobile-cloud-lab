package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"mobile-cloud-lab/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
		"github.com/redis/go-redis/v9"
)

type Handler struct {
	cfg         db.Config
	dbClient    *sql.DB
	redisClient *redis.Client
}

func NewHandler() (*Handler, error) {
	cfg := db.LoadConfig()
	dbClient, err := db.NewPostgresClient(cfg.PostgresURL)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v", err)
		return nil, err
	}
	redisClient := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	return &Handler{cfg, dbClient, redisClient}, nil
}

func (h *Handler) HandleRequest(ctx context.Context, event events.S3Event) (map[string]interface{}, error) {
	log.Printf("POSTGRES_URL: %s", h.cfg.PostgresURL)
	log.Printf("REDIS_ADDR: %s", h.cfg.RedisAddr)

	// Read products.csv
	file, err := os.Open("products.csv")
	if err != nil {
		log.Printf("Failed to open products.csv: %v", err)
		return map[string]interface{}{"error": "Failed to open CSV"}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 4 // name, image, price, qty
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to parse CSV: %v", err)
		return map[string]interface{}{"error": "Failed to parse CSV"}, err
	}

	products := []db.Product{}
	for i, record := range records[1:] { // Skip header
		if len(record) < 4 {
			log.Printf("Invalid record at line %d: %v", i+2, record)
			continue
		}
		price, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Invalid price for product %s at line %d: %v", record[0], i+2, err)
			continue
		}
		qty, err := strconv.Atoi(record[3])
		if err != nil {
			log.Printf("Invalid qty for product %s at line %d: %v", record[0], i+2, err)
			continue
		}
		var image *string
		if record[1] != "" {
			image = &record[1]
		}
		products = append(products, db.Product{
			Name:       record[0],
			Image:      image,
			Price:      price,
			Qty:        qty,
			OutOfStock: qty == 0,
		})
	}
	log.Printf("Parsed %d valid products", len(products))

	// Insert into PostgreSQL
	tx, err := h.dbClient.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return map[string]interface{}{"error": "Transaction start failed"}, err
	}

	for _, product := range products {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO products (name, image, price, qty, out_of_stock) VALUES ($1, $2, $3, $4, $5)",
			product.Name, product.Image, product.Price, product.Qty, product.OutOfStock)
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to insert product %s: %v", product.Name, err)
			return map[string]interface{}{"error": "Database insert failed"}, err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return map[string]interface{}{"error": "Transaction commit failed"}, err
	}
	log.Println("Database transaction committed successfully")

	// Fetch products from PostgreSQL
	rows, err := h.dbClient.QueryContext(ctx, "SELECT id, name, image, price, qty, out_of_stock, created_at, updated_at FROM products")
	if err != nil {
		log.Printf("Failed to fetch products: %v", err)
		return map[string]interface{}{"error": "Database query failed"}, err
	}
	defer rows.Close()

	fetchedProducts := []db.Product{}
	for rows.Next() {
		var p db.Product
		var image sql.NullString
		err := rows.Scan(&p.ID, &p.Name, &image, &p.Price, &p.Qty, &p.OutOfStock, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			log.Printf("Failed to scan product: %v", err)
			return map[string]interface{}{"error": "Database scan failed"}, err
		}
		if image.Valid {
			p.Image = &image.String
		}
		fetchedProducts = append(fetchedProducts, p)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error reading rows: %v", err)
		return map[string]interface{}{"error": "Row iteration failed"}, err
	}
	log.Printf("Fetched %d products from database", len(fetchedProducts))

	// Convert to JSON
	productsJSON, err := json.Marshal(fetchedProducts)
	if err != nil {
		log.Printf("Failed to marshal products: %v", err)
		return map[string]interface{}{"error": "JSON marshal failed"}, err
	}
	log.Printf("Products JSON: %s", string(productsJSON))

	// Set Redis key
	if err = h.redisClient.Set(ctx, "products:all", productsJSON, 3600).Err(); err != nil {
		log.Printf("Failed to set Redis key 'products:all': %v", err)
		return map[string]interface{}{"error": "Redis set failed"}, err
	}
	log.Println("Successfully set Redis key 'products:all'")

	return map[string]interface{}{
		"statusCode": 200,
		"body":       "Products uploaded successfully",
	}, nil
}

func main() {
	handler, err := NewHandler()
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}
	lambda.Start(handler.HandleRequest)
}
