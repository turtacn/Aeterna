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
