// Package mlog provides structured logging functionality using Go's standard slog package.
// It offers context-aware logging with Printf-style formatting and the ability to add attributes to the logging context.
package mlog

import (
	"context"
	"fmt"
)

type debugKey struct{}

// ContextWithDebug creates a new context with debug flag set to true.
func ContextWithDebug(ctx context.Context) context.Context {
	return context.WithValue(ctx, debugKey{}, true)
}

// IsDebugEnabled checks if debug logging is enabled in the context.
func IsDebugEnabled(ctx context.Context) bool {
	if debug, ok := ctx.Value(debugKey{}).(bool); ok {
		return debug
	}
	return false
}

// Debugf logs a message at debug level with printf-style formatting.
// Only logs if debug is enabled in the context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	if IsDebugEnabled(ctx) {
		fmt.Printf(format+"\n", args...)
	}
}

// Infof logs a message at info level with printf-style formatting.
func Infof(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Warnf logs a message at warn level with printf-style formatting.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Errorf logs a message at error level with printf-style formatting.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
