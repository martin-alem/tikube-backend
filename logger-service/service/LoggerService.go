package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"strings"
	"tikube-backend/logger-service/model"
	"tikube-backend/logger-service/repository"
	"tikube-backend/shared/http_error"
	"tikube-backend/shared/kafka_client"
	"tikube-backend/shared/utils"
	"time"
)

type LoggerService struct {
	loggerRepository repository.LoggerRepository
	rdb              *redis.Client
	cache            *cache.Cache
	producer         sarama.AsyncProducer
}

func NewLoggerService(loggerRepository repository.LoggerRepository, rdb *redis.Client, cache *cache.Cache, producer sarama.AsyncProducer) *LoggerService {
	return &LoggerService{loggerRepository: loggerRepository, rdb: rdb, cache: cache}
}

func (ls *LoggerService) ProcessLogs(ctx context.Context, kafkaMessage *sarama.ConsumerMessage) error {
	var receivedLogMsg utils.CreateLogSchema
	err := json.Unmarshal(kafkaMessage.Value, &receivedLogMsg)
	if err != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:SERVICE", err.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, ls.producer)
		return err
	}
	err = ls.loggerRepository.CreateLog(ctx, model.CreateLogSchema(receivedLogMsg))
	if err != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:SERVICE", err.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, ls.producer)
		return err
	}
	return nil
}

func (ls *LoggerService) GetLogs(ctx context.Context, filter utils.LogFilter, pagination utils.Pagination) (*utils.PaginationResult[utils.Log], error) {
	var logTemplate *utils.PaginationResult[utils.Log]
	var repoError error

	key := generateCacheKey(filter, pagination)

	err := ls.cache.Get(ctx, key, &logTemplate)

	if errors.Is(err, cache.ErrCacheMiss) {
		logTemplate, repoError = ls.loggerRepository.GetLogs(ctx, filter, pagination)

		if repoError != nil {
			msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:SERVICE", repoError.Error())
			kafka_client.SendLogToKafka(msg, utils.LoggerTopic, ls.producer)
			return nil, http_error.InternalServerError()
		}

		err := ls.cache.Set(&cache.Item{
			Key:   key,
			Value: logTemplate,
			TTL:   time.Second * 30,
		})
		if err != nil {
			msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:SERVICE", err.Error())
			kafka_client.SendLogToKafka(msg, utils.LoggerTopic, ls.producer)
			return nil, http_error.InternalServerError()
		}
	} else if err != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:SERVICE", err.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, ls.producer)
		return nil, http_error.InternalServerError()
	}

	return logTemplate, nil
}

func generateCacheKey(filter utils.LogFilter, pagination utils.Pagination) string {
	// Combine filter and pagination into a single string
	var sortKey string
	if len(filter.LevelFilter) > 0 {
		sortKey = strings.Join(filter.LevelFilter, "_")
	}

	if filter.DateFilter != nil {
		sortKey += fmt.Sprintf("%s_%s", filter.DateFilter.From, filter.DateFilter.To)
	}

	return fmt.Sprintf("logs_%s_%d_%d", sortKey, pagination.Limit, pagination.Offset)
}
