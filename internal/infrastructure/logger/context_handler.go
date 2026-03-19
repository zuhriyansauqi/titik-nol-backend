package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	RequestContextKey contextKey = "request_id"
)

// ContextHandler wraps a slog.Handler to extract context values into log records.
type ContextHandler struct {
	slog.Handler
}

// Handle extracts the request_id from the context and adds it to the record.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.Handler.Handle(ctx, r)
	}

	if id, ok := ctx.Value(RequestContextKey).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}

	return h.Handler.Handle(ctx, r)
}

// NewContextHandler wraps a slog.Handler with ContextHandler.
func NewContextHandler(h slog.Handler) slog.Handler {
	return &ContextHandler{Handler: h}
}
