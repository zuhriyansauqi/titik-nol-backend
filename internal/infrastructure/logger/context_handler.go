package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	RequestContextKey contextKey = "request_id"
	UserIDContextKey  contextKey = "user_id"
)

// ContextHandler wraps a slog.Handler to extract context values into log records.
type ContextHandler struct {
	inner slog.Handler
}

// Handle extracts the request_id and user_id from the context and adds them to the record.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if id, ok := ctx.Value(RequestContextKey).(string); ok {
			r.AddAttrs(slog.String("request_id", id))
		}
		if uid, ok := ctx.Value(UserIDContextKey).(string); ok {
			r.AddAttrs(slog.String("usr.id", uid))
		}
	}

	return h.inner.Handle(ctx, r)
}

// Enabled delegates to the inner handler.
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// WithAttrs returns a new ContextHandler wrapping the inner handler with additional attributes.
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a new ContextHandler wrapping the inner handler with a group.
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{inner: h.inner.WithGroup(name)}
}

// NewContextHandler wraps a slog.Handler with ContextHandler.
func NewContextHandler(h slog.Handler) slog.Handler {
	return &ContextHandler{inner: h}
}
