package gorm

import (
	"context"
	"errors"
	"time"

	xlogger "github.com/sailxy/x/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Logger = logger.Interface

type CustomLogger struct {
	logger *xlogger.Logger
}

func NewCustomLogger(logger *xlogger.Logger) *CustomLogger {
	return &CustomLogger{
		logger: logger,
	}
}

func (l *CustomLogger) LogMode(lev logger.LogLevel) logger.Interface {
	newLogger := *l
	return &newLogger
}

func (l *CustomLogger) Info(ctx context.Context, msg string, data ...any) {
	l.logger.WithCtx(ctx).Infof(msg, data...)
}

func (l *CustomLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.logger.WithCtx(ctx).Warnf(msg, data...)
}

func (l *CustomLogger) Error(ctx context.Context, msg string, data ...any) {
	l.logger.WithCtx(ctx).Errorf(msg, data...)
}

func (l *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Get SQL execution time.
	elapsed := time.Since(begin)
	t := float64(elapsed.Nanoseconds()) / 1e6
	// Get SQL statements and number of rows affected.
	sql, rows := fc()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.logger.WithCtx(ctx).Errorf("[err: %v] [%.3fms] [rows: %v] %v", err, t, rows, sql)
	} else {
		l.logger.WithCtx(ctx).Infof("[%.3fms] [rows: %v] %v", t, rows, sql)
	}
}
