package logger

import "go.uber.org/zap"

// Logger defines the contract for logging across the application.
// Using an interface allows us to swap the underlying implementation (e.g., for testing).
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	// Structured logging for objects
	DebugObj(msg string, obj any)
	InfoObj(msg string, obj any)

	// Sync flushes any buffered log entries
	Sync() error
}
