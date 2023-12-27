package module

import (
	"context"
	"database/sql"
	"github.com/IBM/sarama"
	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tikube-backend/logger-service/handler"
	"tikube-backend/logger-service/repository"
	"tikube-backend/logger-service/service"
	"tikube-backend/shared/kafka_client"
	"tikube-backend/shared/middleware"
	"tikube-backend/shared/utils"
	"time"
)

func LoggerModule(router *mux.Router, db *sql.DB, redisCache *cache.Cache, rdb *redis.Client, limiter *redis_rate.Limiter, consumer sarama.ConsumerGroup, producer sarama.AsyncProducer) {

	loggerRepository := repository.NewLoggerRepository(db, producer)
	loggerService := service.NewLoggerService(loggerRepository, rdb, redisCache, producer)
	loggerHandler := handlers.NewLoggerController(loggerService, producer)

	loggerRouter := router.PathPrefix("/logger").Subrouter()

	loggerRouter.HandleFunc("/logs",
		shared_middleware.ErrorHandlerMiddleware(shared_middleware.ChainMiddlewares(
			loggerHandler.GetLogs,
			shared_middleware.CorsMiddleware,
			shared_middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   1000,
				Burst:  100,
				Period: time.Minute * 1,
			}),
			shared_middleware.LoggingMiddleware))).Methods("GET")

	ctx, cancel := context.WithCancel(context.Background())

	// Start consuming messages in a goroutine for capturing log events
	go func() {
		defer func(consumer sarama.ConsumerGroup) {
			err := consumer.Close()
			if err != nil {
				return
			}
		}(consumer)
		for {
			// Use the cancellable context with the Kafka consumer
			if err := consumer.Consume(ctx, []string{utils.LoggerTopic}, kafka_client.ConsumerGroupHandler{Processor: loggerService.ProcessLogs}); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				// Context is canceled, exit the goroutine
				log.Println("Consumer context is cancelled")
				return
			}
		}
	}()

	// Listen for shutdown signals
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		cancel()
	}()
}
