package main

import (
	"context"
	"errors"
	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"tikube-backend/logger-service/module"
	"tikube-backend/shared/kafka_client"
	"tikube-backend/shared/mysql"
	"tikube-backend/shared/redis"
	"tikube-backend/shared/utils"
	"time"
)

func main() {

	dotErr := godotenv.Load()
	if dotErr != nil {
		log.Fatalf("Error loading .env file: %v", dotErr)
	}

	r := mux.NewRouter()

	db := mysql.ConnectToDatabase()

	dir, _ := os.Getwd()

	sqlBytes, readErr := os.ReadFile(filepath.Join(dir, "/cmd/sql/setup.sql"))
	if readErr != nil {
		log.Fatalf("Error reading SQL file: %v", readErr)
	}
	sqlScript := string(sqlBytes)

	// Execute the SQL script
	_, execErr := db.Exec(sqlScript)
	if execErr != nil {
		log.Fatalf("Error executing SQL script: %v", execErr)
	}

	// Initialize Kafka producer and consumer
	producer, consumer, err := kafka_client.KafkaClient(utils.LoggerGroupId)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %s", err)
	}

	// Initialize Redis
	rdb := redis.ConnectToRedis()

	// Setup Redis Cache
	redisCache := cache.New(&cache.Options{
		Redis: rdb,
	})

	// Setup Redis RateLimiter
	rateLimiter := redis_rate.NewLimiter(rdb)

	//Mounting modules
	module.LoggerModule(r, db, redisCache, rdb, rateLimiter, consumer, producer)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 2^20 shifting 1 left by 20 = 1,048,576
	}

	log.Println("Server listening on port 8080")
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Setting up a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbErr := db.Close()
	if dbErr != nil {
		log.Fatal("Error disconnecting from mysql database: ", dbErr)
	}

	err = rdb.Close()
	if err != nil {
		log.Fatal("Error disconnecting from redis: ", err)
	}

	err = producer.Close()
	if err != nil {
		log.Fatal("Error disconnecting from kafka: ", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
