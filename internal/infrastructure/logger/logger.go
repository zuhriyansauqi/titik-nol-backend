package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
)

// Initialize sets up the global slog logger based on config.
// Base attributes (service, env) are attached to every log line,
// which enables Datadog's Unified Service Tagging out of the box.
func Initialize(cfg *config.Config) {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: strings.ToLower(cfg.AppEnv) != "production",
	}

	var handler slog.Handler
	if strings.ToLower(cfg.LogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Attach service-level attributes for Datadog Unified Service Tagging.
	// These appear on every log line without manual repetition.
	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", cfg.AppName),
		slog.String("env", cfg.AppEnv),
	})

	logger := slog.New(NewContextHandler(handler))
	slog.SetDefault(logger)
}
