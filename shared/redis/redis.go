package redis

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
)

func ConnectToRedis() *redis.Client {
	// Define Redis options
	options := &redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PWD"),
		Username: os.Getenv("REDIS_USERNAME"),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	// Create a new Redis client with the options
	rdb := redis.NewClient(options)

	// Test the connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis server: %v", err)
	}

	return rdb
}
