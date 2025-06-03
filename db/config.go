package db

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresURL   string
	RedisAddr     string
	RedisPassword string
	S3Bucket      string
	AWSRegion     string
	IsLocal       bool
}

func LoadConfig() Config {
	// Load .env file
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("Failed to load .env file: %v", err))
	}

	// Construct PostgreSQL URL
	postgresURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Construct Redis address
	redisAddr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))

	cfg := Config{
		PostgresURL:   postgresURL,
		RedisAddr:     redisAddr,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		S3Bucket:      os.Getenv("S3_BUCKET"),
		AWSRegion:     os.Getenv("AWS_REGION"),
		IsLocal:       strings.ToLower(os.Getenv("NODE_ENV")) == "local",
	}

	// Fallback defaults for local development
	if cfg.IsLocal {
		cfg.PostgresURL = "postgres://postgres:root@host.docker.internal:5432/product_db?sslmode=disable"
		cfg.RedisAddr = "host.docker.internal:6379"
	}
	if cfg.S3Bucket == "" {
		cfg.S3Bucket = "product-data-bucket"
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = "us-east-1"
	}

	return cfg
}