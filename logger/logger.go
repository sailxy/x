package logger

import (
	"context"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type traceIDKey struct{}

type Config struct {
	Skip    int
	AppName string

	// Log file config.
	Path       string
	MaxSize    int // megabytes
	MaxAge     int // days
	MaxBackups int
	Compress   bool
}

type Logger struct {
	instance *zap.SugaredLogger
}

func New(c Config) (*Logger, error) {
	// Writing logs to rolling files.
	lj := &lumberjack.Logger{
		Filename:   c.Path,
		MaxSize:    c.MaxSize, // megabytes
		MaxAge:     c.MaxAge,  //days
		MaxBackups: c.MaxBackups,
		Compress:   c.Compress, // disabled by default
	}

	// Output logs to both the terminal and the file.
	writer := io.MultiWriter(lj, os.Stdout)
	writeSyncer := zapcore.AddSync(writer)

	// Set the log time format to human-readable time.
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	// Set the log level information to uppercase.
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	z := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(c.Skip),
	)

	return &Logger{
		instance: z.Sugar().With("app", c.AppName),
	}, nil
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	traceID := ctx.Value(traceIDKey{})
	return &Logger{
		instance: l.instance.With("trace_id", traceID),
	}
}

func (l *Logger) Debug(args ...any) {
	l.instance.Debugln(args...)
}

func (l *Logger) Debugf(template string, args ...any) {
	l.instance.Debugf(template, args...)
}

func (l *Logger) Info(args ...any) {
	l.instance.Infoln(args...)
}

func (l *Logger) Infof(template string, args ...any) {
	l.instance.Infof(template, args...)
}

func (l *Logger) Warn(args ...any) {
	l.instance.Warnln(args...)
}

func (l *Logger) Warnf(template string, args ...any) {
	l.instance.Warnf(template, args...)
}

func (l *Logger) Error(args ...any) {
	l.instance.Errorln(args...)
}

func (l *Logger) Errorf(template string, args ...any) {
	l.instance.Errorf(template, args...)
}

func (l *Logger) Fatal(args ...any) {
	l.instance.Fatalln(args...)
}

func (l *Logger) Fatalf(template string, args ...any) {
	l.instance.Fatalf(template, args...)
}

func (l *Logger) Panic(args ...any) {
	l.instance.Panicln(args...)
}

func (l *Logger) Panicf(template string, args ...any) {
	l.instance.Panicf(template, args...)
}

func SetTraceID(ctx context.Context, traceID any) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}
