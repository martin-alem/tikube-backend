package utils

import (
	"net/http"
	"time"
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

const LoggerGroupId = "logger-consumers"
const LoggerTopic = "log_events"

const MaxPayloadSize int64 = 10_485_760 // 10MB
type PayloadKey struct{}

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type Log struct {
	Id        int64     `json:"id"`
	LogLevel  LogLevel  `json:"logLevel"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PaginationResult[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

type DateFilterRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type LogFilter struct {
	LevelFilter []string         `json:"levelFilter"`
	DateFilter  *DateFilterRange `json:"dateFilter"`
}

type CreateLogSchema struct {
	LogLevel LogLevel
	Source   string
	Message  string
}

type Middleware func(HTTPHandler) HTTPHandler

type HTTPHandler func(http.ResponseWriter, *http.Request) error

type ValidatableSchema interface {
	Validate() error
}
