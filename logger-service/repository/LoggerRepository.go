package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/IBM/sarama"
	"strings"
	"tikube-backend/logger-service/model"
	"tikube-backend/shared/kafka_client"
	"tikube-backend/shared/utils"
)

type LoggerRepository interface {
	CreateLog(ctx context.Context, log model.CreateLogSchema) error
	GetLogs(ctx context.Context, filter utils.LogFilter, pagination utils.Pagination) (*utils.PaginationResult[utils.Log], error)
}

type SQLLoggerRepository struct {
	db       *sql.DB
	producer sarama.AsyncProducer
}

func NewLoggerRepository(db *sql.DB, producer sarama.AsyncProducer) LoggerRepository {
	return &SQLLoggerRepository{db: db, producer: producer}
}

func (repo *SQLLoggerRepository) CreateLog(ctx context.Context, log model.CreateLogSchema) error {

	query := `INSERT INTO logs (logLevel, source, message) VALUES (?, ?, ?)`

	result, err := repo.db.ExecContext(ctx, query, log.LogLevel, log.Source, log.Message)
	if err != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return err
	}

	selectStmt := `SELECT id, logLevel, source, message, createdAt, updatedAt FROM logs WHERE id = ?`
	var insertedLog utils.Log
	err = repo.db.QueryRowContext(ctx, selectStmt, lastInsertedId).Scan(&insertedLog.Id, &insertedLog.LogLevel, &insertedLog.Source, &insertedLog.Message, &insertedLog.CreatedAt, &insertedLog.UpdatedAt)
	if err != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return err
	}

	return nil
}

func (repo *SQLLoggerRepository) GetLogs(ctx context.Context, filter utils.LogFilter, pagination utils.Pagination) (*utils.PaginationResult[utils.Log], error) {
	// Start building the query
	baseQuery := "SELECT id, logLevel, source, message, createdAt, updatedAt FROM logs"
	whereBaseQuery := ""
	var queryParams []any
	var countQueryParams []any // An extra slice is needed for count query because limit and offset are omitted when counting. QueryContext is strict about the number of args passed for a query

	//Dynamically constructing WHERE statement for log level using ? as query parameters.
	//Also ensuring that OR is not appended at the end of the query.
	lenOfLevelFilter := len(filter.LevelFilter)
	if lenOfLevelFilter > 0 {
		whereBaseQuery += " WHERE ( "
		for i, v := range filter.LevelFilter {
			queryParams = append(queryParams, strings.ToUpper(v))
			countQueryParams = append(countQueryParams, strings.ToUpper(v))
			whereBaseQuery += fmt.Sprintf(" logLevel = %s ", "?")
			if i != lenOfLevelFilter-1 {
				whereBaseQuery += " OR "
			}
		}

		whereBaseQuery += " ) "
	}

	//Date filter
	if filter.DateFilter != nil {
		queryParams = append(queryParams, filter.DateFilter.From, filter.DateFilter.To)
		countQueryParams = append(countQueryParams, filter.DateFilter.From, filter.DateFilter.To)
		if lenOfLevelFilter > 0 {
			whereBaseQuery += fmt.Sprintf(" AND ( createdAt >= %s AND createdAt <= %s )", "?", "?")
		} else {
			whereBaseQuery += fmt.Sprintf(" WHERE createdAt >= %s AND createdAt <= %s ", "?", "?")
		}
	}

	//Add where query
	baseQuery += whereBaseQuery

	// Add pagination
	queryParams = append(queryParams, pagination.Limit, pagination.Offset)
	baseQuery += fmt.Sprintf(" LIMIT %s OFFSET %s", "?", "?")

	// Execute base query
	baseQueryRows, baseQueryErr := repo.db.QueryContext(ctx, baseQuery, queryParams...)

	//Close base query row operation
	defer func() {
		err := baseQueryRows.Close()
		if err != nil {
			msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
			kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		}
	}()

	if baseQueryErr != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", baseQueryErr.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return nil, baseQueryErr
	}

	var logs []utils.Log = nil
	for baseQueryRows.Next() {
		var dLog utils.Log
		if err := baseQueryRows.Scan(&dLog.Id, &dLog.LogLevel, &dLog.Source, &dLog.Message, &dLog.CreatedAt, &dLog.UpdatedAt); err != nil {
			msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
			kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
			return nil, err
		}
		logs = append(logs, dLog)
	}

	// Check for any errors encountered during base query iteration
	if baseQueryIterErr := baseQueryRows.Err(); baseQueryIterErr != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", baseQueryIterErr.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return nil, baseQueryIterErr
	}

	totalResultQuery := "SELECT COUNT(*) AS Total FROM logs" + whereBaseQuery

	//Execute count query
	countQueryRows, countQueryErr := repo.db.QueryContext(ctx, totalResultQuery, countQueryParams...)

	//Close count query row operation
	defer func() {
		err := countQueryRows.Close()
		if err != nil {
			msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
			kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		}
	}()

	if countQueryErr != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", countQueryErr.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return nil, countQueryErr
	}

	var totalResult int
	for countQueryRows.Next() {
		if err := countQueryRows.Scan(&totalResult); err != nil {
			msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", err.Error())
			kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
			return nil, err
		}
	}
	// Check for any errors encountered during count query iteration
	if countQueryIterErr := countQueryRows.Err(); countQueryIterErr != nil {
		msg := utils.CreateSerializedLog(utils.FATAL, "LOGGER:REPOSITORY", countQueryIterErr.Error())
		kafka_client.SendLogToKafka(msg, utils.LoggerTopic, repo.producer)
		return nil, countQueryIterErr
	}

	//Ensures that Golang marshal an empty slice into an empty array ([]) in json
	if len(logs) <= 0 {
		logs = []utils.Log{}
	}

	return &utils.PaginationResult[utils.Log]{Data: logs, Total: totalResult}, nil
}
