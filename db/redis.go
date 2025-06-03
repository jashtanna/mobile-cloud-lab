package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		panic(err)
	}
	return client
}
