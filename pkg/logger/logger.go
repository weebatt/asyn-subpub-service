package logger

import (
	"context"
	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

const (
	key = "logger"
)

func New(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, key, &Logger{logger: logger}), err
}

func GetLoggerFromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(key).(*Logger)
	if !ok {
		return &Logger{zap.NewNop()}
	}
	return logger
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}
