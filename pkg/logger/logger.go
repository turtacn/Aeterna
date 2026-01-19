package logger

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

var Log Logger

func InitLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
		// Add source file info for better debugging
		AddSource: true,
	}
	// Use JSON handler for cloud-native observability
	handler := slog.NewJSONHandler(os.Stdout, opts)
	l := slog.New(handler)

	Log = &wrapper{l: l}
}

type wrapper struct {
	l *slog.Logger
}

func (w *wrapper) Debug(msg string, args ...any) { w.l.Debug(msg, args...) }
func (w *wrapper) Info(msg string, args ...any)  { w.l.Info(msg, args...) }
func (w *wrapper) Warn(msg string, args ...any)  { w.l.Warn(msg, args...) }
func (w *wrapper) Error(msg string, args ...any) { w.l.Error(msg, args...) }
func (w *wrapper) With(args ...any) Logger       { return &wrapper{l: w.l.With(args...)} }

// Personal.AI order the ending
