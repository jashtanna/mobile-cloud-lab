package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestGetAllHandler_RedisProducts(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("NODE_ENV", "local")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	// Set up test data in Redis
	err := rdb.Set(context.Background(), "product:1", `{"id":"1","name":"Test Product","price":10.99,"qty":100}`, 0).Err()
	if err != nil {
		t.Fatalf("Failed to set Redis test data: %v", err)
	}

	resp, err := getAllHandler(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var products []Product
	if err := json.Unmarshal([]byte(resp.Body), &products); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(products) != 1 || products[0].Name != "Test Product" {
		t.Errorf("Expected 1 product with name 'Test Product', got %v", products)
	}
	if resp.Headers["Cache-Control"] != "public, max-age=300" {
		t.Errorf("Expected Cache-Control header, got %v", resp.Headers)
	}
}
