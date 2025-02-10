package gorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sailxy/x/logger"
	"github.com/stretchr/testify/assert"
)

func newCustomLogger() (*CustomLogger, error) {
	logger, err := logger.New(logger.Config{})
	if err != nil {
		return nil, err
	}

	return NewCustomLogger(logger), nil
}

func TestLogPrint(t *testing.T) {
	cl, err := newCustomLogger()
	assert.NoError(t, err)

	ctx := context.Background()
	ctx = logger.SetTraceID(ctx, "123456")
	cl.Info(ctx, "test info log with number %v", 0)
	cl.Warn(ctx, "test warn log with text %v", "hello")
	cl.Error(ctx, "test error log with error %v", errors.New("test error"))
	cl.Trace(ctx, time.Now(), func() (sql string, rowsAffected int64) {
		return "select * from user", 1
	}, nil)
	cl.Trace(ctx, time.Now(), func() (sql string, rowsAffected int64) {
		return "select * from user", 1
	}, errors.New("sql error"))
}
