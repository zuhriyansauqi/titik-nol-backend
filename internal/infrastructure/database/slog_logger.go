package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SlogLogger struct {
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l *SlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *SlogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		slog.InfoContext(ctx, msg, "data", data)
	}
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		slog.WarnContext(ctx, msg, "data", data)
	}
}

func (l *SlogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		slog.ErrorContext(ctx, msg, "data", data)
	}
}

func (l *SlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		slog.ErrorContext(ctx, "gorm trace error",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
			"err", err,
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slog.WarnContext(ctx, "gorm trace slow",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	case l.LogLevel >= logger.Info:
		sql, rows := fc()
		slog.InfoContext(ctx, "gorm trace",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	}
}
