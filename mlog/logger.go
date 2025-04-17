// Package mlog provides structured logging functionality using Go's standard slog package.
// It offers context-aware logging with Printf-style formatting and the ability to add attributes to the logging context.
package mlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// SetDefaultLogger sets the default logger used when no logger is found in the context.
func SetDefaultLogger(logger *slog.Logger) {
	defaultLogger = logger
}

// GetDefaultLogger returns the current default logger.
func GetDefaultLogger() *slog.Logger {
	return defaultLogger
}

// Debugf logs a message at debug level with printf-style formatting.
// It retrieves the logger from the context and includes any attributes stored in the context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	} else {
		logger.Debug(msg)
	}
}

// Infof logs a message at info level with printf-style formatting.
// It retrieves the logger from the context and includes any attributes stored in the context.
func Infof(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	} else {
		logger.Info(msg)
	}
}

// Warnf logs a message at warn level with printf-style formatting.
// It retrieves the logger from the context and includes any attributes stored in the context.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	} else {
		logger.Warn(msg)
	}
}

// Errorf logs a message at error level with printf-style formatting.
// It retrieves the logger from the context and includes any attributes stored in the context.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	} else {
		logger.Error(msg)
	}
}
