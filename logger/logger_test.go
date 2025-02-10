package logger

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const logDir = "./logs"
const logPath = "./logs/test.log"

func TestInfo(t *testing.T) {
	logger, err := New(Config{
		Skip:       1,
		AppName:    "test",
		Path:       logPath,
		MaxSize:    500,
		MaxAge:     30,
		MaxBackups: 30,
	})
	assert.NoError(t, err)
	defer os.RemoveAll(logDir)

	logger.Info("hello", "world")
}

func TestLogWithTraceID(t *testing.T) {
	logger, err := New(Config{
		Skip:       1,
		AppName:    "test",
		Path:       logPath,
		MaxSize:    500,
		MaxAge:     30,
		MaxBackups: 30,
	})
	assert.NoError(t, err)
	defer os.RemoveAll(logDir)

	ctx := SetTraceID(context.Background(), "123456789")
	logger.WithCtx(ctx).Info("message with trace id")
}
