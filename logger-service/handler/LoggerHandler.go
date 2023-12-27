package handlers

import (
	"github.com/IBM/sarama"
	"net/http"
	"strconv"
	"strings"
	"tikube-backend/logger-service/service"
	"tikube-backend/shared/http_error"
	"tikube-backend/shared/utils"
)

type LoggerHandler struct {
	loggerService *service.LoggerService
	producer      sarama.AsyncProducer
}

func NewLoggerController(loggerService *service.LoggerService, producer sarama.AsyncProducer) *LoggerHandler {
	return &LoggerHandler{loggerService: loggerService}
}

func (lc *LoggerHandler) GetLogs(w http.ResponseWriter, r *http.Request) error {
	const defaultLimit int = 10
	const defaultOffset int = 0
	var filter utils.LogFilter

	query := r.URL.Query()

	limit := defaultLimit
	offset := defaultOffset

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit >= 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	offset = offset * limit

	if filterStr := query.Get("level_filter"); filterStr != "" {
		filter.LevelFilter = strings.Split(filterStr, ",")
	}

	//Date filter
	if dateFilterStr := query.Get("date_filter"); dateFilterStr != "" {
		dateFilterSlice := strings.Split(dateFilterStr, ",")
		if len(dateFilterSlice) == 2 {
			// Optionally add more validation for date format
			filter.DateFilter = &utils.DateFilterRange{From: dateFilterSlice[0], To: dateFilterSlice[1]}
		} else {
			filter.DateFilter = nil
		}
	}

	logs, err := lc.loggerService.GetLogs(r.Context(), filter, utils.Pagination{Limit: limit, Offset: offset})
	if err != nil {
		return http_error.InternalServerError()
	}
	//Return an empty json array instead of nil
	if logs == nil {
		logs = &utils.PaginationResult[utils.Log]{Data: []utils.Log{}, Total: 0}
	}
	return utils.JSONResponse(w, http.StatusOK, logs)
}
