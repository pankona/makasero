package mlog

import (
	"context"
	"log/slog"
)

type ContextAwareHandler struct {
	handler slog.Handler
}

func NewContextAwareHandler(h slog.Handler) *ContextAwareHandler {
	return &ContextAwareHandler{handler: h}
}

func (h *ContextAwareHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		for _, attr := range attrs {
			r.AddAttrs(attr)
		}
	}
	
	return h.handler.Handle(ctx, r)
}

func (h *ContextAwareHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewContextAwareHandler(h.handler.WithAttrs(attrs))
}

func (h *ContextAwareHandler) WithGroup(name string) slog.Handler {
	return NewContextAwareHandler(h.handler.WithGroup(name))
}

func (h *ContextAwareHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}
