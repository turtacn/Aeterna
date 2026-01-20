package logger

import (
	"testing"
)

func TestLogger_DefaultInitialization(t *testing.T) {
	// Log should be initialized by default and not panic
	if Log == nil {
		t.Fatal("Log should not be nil by default")
	}

	// Should not panic
	Log.Info("Testing default logger")
}

func TestInitLogger(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			InitLogger(level)
			if Log == nil {
				t.Fatalf("Log should not be nil after InitLogger(%s)", level)
			}
			Log.Debug("test debug")
			Log.Info("test info")
			Log.Warn("test warn")
			Log.Error("test error")
		})
	}
}

func TestLogger_With(t *testing.T) {
	InitLogger("info")
	loggerWith := Log.With("key", "value")
	if loggerWith == nil {
		t.Fatal("Log.With should not return nil")
	}
	loggerWith.Info("test info with args")
}
