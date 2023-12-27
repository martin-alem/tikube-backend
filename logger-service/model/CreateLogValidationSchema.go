package model

import (
	"errors"
	"strings"
	"tikube-backend/shared/utils"
)

type CreateLogSchema struct {
	LogLevel utils.LogLevel `json:"logLevel"`
	Source   string         `json:"source"`
	Message  string         `json:"message"`
}

func NewCreateLogSchema() CreateLogSchema {
	return CreateLogSchema{}
}

func (s CreateLogSchema) Validate() error {
	// Validate LogLevel
	if strings.ToUpper(string(s.LogLevel)) != "INFO" && strings.ToUpper(string(s.LogLevel)) != "WARN" && strings.ToUpper(string(s.LogLevel)) != "ERROR" && strings.ToUpper(string(s.LogLevel)) != "FATAL" {
		return errors.New("invalid log level")
	}

	// Validate Source
	trimmedSource := strings.TrimSpace(s.Source)
	if trimmedSource == "" {
		return errors.New("source is required")
	}

	// Validate Message
	trimmedMessage := strings.TrimSpace(s.Message)
	if trimmedMessage == "" {
		return errors.New("message is required")
	}

	return nil
}
