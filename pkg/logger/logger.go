package logger

import (
	"log/slog"
	"os"
)

// Logger defines the interface for logging in the Aeterna system.
// It provides standard logging levels and a mechanism to add structured context.
type Logger interface {
	// Debug logs a message at the debug level.
	Debug(msg string, args ...any)
	// Info logs a message at the info level.
	Info(msg string, args ...any)
	// Warn logs a message at the warning level.
	Warn(msg string, args ...any)
	// Error logs a message at the error level.
	Error(msg string, args ...any)
	// With returns a new Logger with the given structured context added.
	With(args ...any) Logger
}

// Log is the global logger instance used throughout the application.
// It is initialized with a default JSON handler pointing to stdout.
var Log Logger = &wrapper{l: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}))}

// InitLogger initializes the global Log instance with the specified logging level.
// Supported levels are "debug", "info", "warn", and "error".
// It uses a JSON handler and includes source file information in the output.
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
