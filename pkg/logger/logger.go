package logger

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

type logger struct {
	slog *slog.Logger
}

func NewLogger() Logger {
	return &logger{
		slog: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

func (l *logger) Debug(msg string, args ...interface{}) {
	l.slog.Debug(msg, args...)
}

func (l *logger) Info(msg string, args ...interface{}) {
	l.slog.Info(msg, args...)
}

func (l *logger) Warn(msg string, args ...interface{}) {
	l.slog.Warn(msg, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	l.slog.Error(msg, args...)
}

func (l *logger) Fatal(msg string, args ...interface{}) {
	l.slog.Error(msg, args...)
	os.Exit(1)
}
