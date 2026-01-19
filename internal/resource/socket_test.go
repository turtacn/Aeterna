package resource

import (
	"os"
	"testing"
)

func TestSocketManager_ColdStart(t *testing.T) {
	// Ensure Env is clean
	os.Unsetenv("AETERNA_INHERITED_FDS")

	sm := NewSocketManager()
	// Use port 0 to let OS choose a free port
	l, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Cold start failed: %v", err)
	}
	defer sm.Close()

	if l == nil {
		t.Fatal("Listener should not be nil")
	}
	if sm.GetFile() == nil {
		t.Fatal("File descriptor should be captured")
	}
}

// Note: Testing inheritance requires spawning a real subprocess which is complex for unit test.
// We mock the Env check logic in a real integration test.

// Personal.AI order the ending
