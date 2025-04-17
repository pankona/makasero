package mlog

import (
	"context"
	"log/slog"
)

// ContextAwareHandler is a custom slog.Handler that extracts attributes from the context
// and adds them to log records.
type ContextAwareHandler struct {
	handler slog.Handler
}

// NewContextAwareHandler creates a new handler that wraps an existing slog.Handler
// and adds context attributes to log records.
func NewContextAwareHandler(h slog.Handler) *ContextAwareHandler {
	return &ContextAwareHandler{handler: h}
}

// Handle processes a log record by adding any attributes from the context
// before passing it to the underlying handler.
func (h *ContextAwareHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		for _, attr := range attrs {
			r.AddAttrs(attr)
		}
	}
	
	return h.handler.Handle(ctx, r)
}

// WithAttrs returns a new Handler whose attributes include both the receiver's
// attributes and the arguments.
func (h *ContextAwareHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewContextAwareHandler(h.handler.WithAttrs(attrs))
}

// WithGroup returns a new Handler with the given group name added to the
// receiver's existing groups.
func (h *ContextAwareHandler) WithGroup(name string) slog.Handler {
	return NewContextAwareHandler(h.handler.WithGroup(name))
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *ContextAwareHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}
