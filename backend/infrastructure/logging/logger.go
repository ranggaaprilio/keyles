package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Logger is a structured slog wrapper with level filtering
type Logger struct {
	inner *slog.Logger
	level slog.Level
}

// NewLogger creates a new structured logger with the specified log level
func NewLogger(level string) *Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	inner := slog.New(handler)

	return &Logger{
		inner: inner,
		level: logLevel,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.inner.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
}

// With returns a logger with the given attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		inner: l.inner.With(args...),
		level: l.level,
	}
}
