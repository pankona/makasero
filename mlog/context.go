package mlog

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type loggerKey struct{}
type logAttrKey struct{}

// LoggerFromContext retrieves a logger from the context.
// If no logger is found in the context, it returns the default logger.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return defaultLogger
}

// ContextWithLogger creates a new context with the specified logger.
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// ContextWithDebug creates a new context with a logger configured for debug level logging.
func ContextWithDebug(ctx context.Context) context.Context {
	// Create a new handler with debug level
	debugHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	debugLogger := slog.New(debugHandler)
	return ContextWithLogger(ctx, debugLogger)
}

// ContextWithLogLevel creates a new context with a logger configured for the specified log level.
func ContextWithLogLevel(ctx context.Context, level slog.Level) context.Context {
	levelHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	newLogger := slog.New(levelHandler)
	return ContextWithLogger(ctx, newLogger)
}

// ContextWithAttr adds an attribute to the context for logging.
// The attribute will be included in all log messages created with this context.
func ContextWithAttr(ctx context.Context, key string, value interface{}) context.Context {
	var attrs []slog.Attr
	if existingAttrs, ok := ctx.Value(logAttrKey{}).([]slog.Attr); ok {
		attrs = existingAttrs
	}
	
	var attr slog.Attr
	switch v := value.(type) {
	case string:
		attr = slog.String(key, v)
	case int:
		attr = slog.Int(key, v)
	case bool:
		attr = slog.Bool(key, v)
	case float64:
		attr = slog.Float64(key, v)
	case time.Time:
		attr = slog.Time(key, v)
	case time.Duration:
		attr = slog.Duration(key, v)
	default:
		attr = slog.Any(key, v)
	}
	
	attrs = append(attrs, attr)
	return context.WithValue(ctx, logAttrKey{}, attrs)
}

// ContextWithSessionID adds a session_id attribute to the context for logging.
// This is a convenience function for ContextWithAttr with the key "session_id".
func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return ContextWithAttr(ctx, "session_id", sessionID)
}

// AttrsFromContext retrieves the logging attributes from the context.
// Returns nil if no attributes are found.
func AttrsFromContext(ctx context.Context) []slog.Attr {
	if attrs, ok := ctx.Value(logAttrKey{}).([]slog.Attr); ok {
		return attrs
	}
	return nil
}
