package mlog

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type loggerKey struct{}
type logAttrKey struct{}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return defaultLogger
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func ContextWithDebug(ctx context.Context) context.Context {
	debugHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	debugLogger := slog.New(debugHandler)
	return ContextWithLogger(ctx, debugLogger)
}

func ContextWithLogLevel(ctx context.Context, level slog.Level) context.Context {
	levelHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	newLogger := slog.New(levelHandler)
	return ContextWithLogger(ctx, newLogger)
}

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

func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return ContextWithAttr(ctx, "session_id", sessionID)
}

func AttrsFromContext(ctx context.Context) []slog.Attr {
	if attrs, ok := ctx.Value(logAttrKey{}).([]slog.Attr); ok {
		return attrs
	}
	return nil
}
