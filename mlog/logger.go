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

func SetDefaultLogger(logger *slog.Logger) {
	defaultLogger = logger
}

func GetDefaultLogger() *slog.Logger {
	return defaultLogger
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	} else {
		logger.Debug(msg)
	}
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	} else {
		logger.Info(msg)
	}
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	} else {
		logger.Warn(msg)
	}
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger := LoggerFromContext(ctx)
	
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	} else {
		logger.Error(msg)
	}
}
