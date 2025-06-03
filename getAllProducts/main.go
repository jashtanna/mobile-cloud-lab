package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

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
		return nil, err
	}
	redisClient := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	return &Handler{cfg, dbClient, redisClient}, nil
}

func (h *Handler) HandleRequest(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	productsJSON, err := h.redisClient.Get(ctx, "products:all").Result()
	if err == redis.Nil {
		rows, err := h.dbClient.QueryContext(ctx, `
            SELECT id, name, image, price, qty, out_of_stock, created_at, updated_at
            FROM products
        `)
		if err != nil {
			log.Printf("Failed to query database: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       `{"error": "Internal server error"}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			}, err
		}
		defer rows.Close()

		var products []db.Product
		for rows.Next() {
			var p db.Product
			var image sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &image, &p.Price, &p.Qty, &p.OutOfStock, &p.CreatedAt, &p.UpdatedAt); err != nil {
				log.Printf("Failed to scan row: %v", err)
				continue
			}
			if image.Valid {
				p.Image = &image.String
			}
			products = append(products, p)
		}

		productsJSON, err := json.Marshal(products)
		if err != nil {
			log.Printf("Failed to marshal products: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       `{"error": "Internal server error"}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			}, err
		}
		err = h.redisClient.Set(ctx, "products:all", productsJSON, 3600).Err()
		if err != nil {
			log.Printf("Failed to cache in Redis: %v", err)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(productsJSON),
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Cache-Control":               "max-age=300",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	} else if err != nil {
		log.Printf("Redis error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Internal server error"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       productsJSON,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Cache-Control":               "max-age=300",
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	handler, err := NewHandler()
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}
	lambda.Start(handler.HandleRequest)
}
