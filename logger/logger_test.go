package logger

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestInfo(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "test.log")
	logger, err := New(Config{
		Skip:       1,
		AppName:    "test",
		Path:       logPath,
		MaxSize:    500,
		MaxAge:     30,
		MaxBackups: 30,
	})
	assert.NoError(t, err)

	logger.Info("hello", "world")
}

func TestLogWithTraceID(t *testing.T) {
	ctx := SetTraceID(context.Background(), "123456789")
	output := captureStdout(t, func() {
		logPath := filepath.Join(t.TempDir(), "test.log")
		logger, err := New(Config{
			Skip:       1,
			AppName:    "test",
			Path:       logPath,
			MaxSize:    500,
			MaxAge:     30,
			MaxBackups: 30,
		})
		assert.NoError(t, err)
		logger.WithCtx(ctx).Info("message with trace id")
	})
	assert.Contains(t, output, `"trace_id":"123456789"`)
}

func TestLogWithOTelTraceID(t *testing.T) {
	traceID, err := trace.TraceIDFromHex("1234567890abcdef1234567890abcdef")
	assert.NoError(t, err)
	spanID, err := trace.SpanIDFromHex("1234567890abcdef")
	assert.NoError(t, err)

	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	output := captureStdout(t, func() {
		logPath := filepath.Join(t.TempDir(), "test.log")
		logger, err := New(Config{
			Skip:       1,
			AppName:    "test",
			Path:       logPath,
			MaxSize:    500,
			MaxAge:     30,
			MaxBackups: 30,
		})
		assert.NoError(t, err)
		logger.WithCtx(ctx).Info("message with trace id from otel")
	})
	assert.Contains(t, output, `"trace_id":"1234567890abcdef1234567890abcdef"`)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	err = w.Close()
	assert.NoError(t, err)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	return strings.TrimSpace(buf.String())
}
