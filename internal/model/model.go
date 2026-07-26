package model

import "encoding/json"

type LogLevel string

const (
	ErrorLevel LogLevel = "ERROR"
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
)

type LogEntry struct {
	Level       LogLevel        `json:"level"`
	Timestamp   string          `json:"timestamp"`
	ServiceName string          `json:"service_name"`
	Message     json.RawMessage `json:"message"`
}
