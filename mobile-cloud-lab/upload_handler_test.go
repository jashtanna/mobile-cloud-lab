package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestUploadHandler_ValidCSV(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("NODE_ENV", "local")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "product_db")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "root")

	// Create test CSV
	os.WriteFile("products.csv", []byte("id,name,image,price,qty\n1,Test Product,http://test.jpg,10.99,100"), 0644)
	defer os.Remove("products.csv")

	resp, err := uploadHandler(context.Background(), events.S3Event{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if result["message"] != "Products uploaded successfully" {
		t.Errorf("Expected message 'Products uploaded successfully', got %v", result["message"])
	}
	if result["count"] != float64(1) { // JSON numbers are parsed as float64
		t.Errorf("Expected count 1, got %v", result["count"])
	}
}

func TestUploadHandler_InvalidCSV(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("NODE_ENV", "local")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "product_db")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "root")

	// Create invalid CSV
	os.WriteFile("products.csv", []byte("invalid"), 0644)
	defer os.Remove("products.csv")

	resp, err := uploadHandler(context.Background(), events.S3Event{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if result["message"] != "No valid products in CSV" {
		t.Errorf("Expected message 'No valid products in CSV', got %v", result["message"])
	}
}