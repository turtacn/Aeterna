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

func TestSocketManager_CloseResets(t *testing.T) {
	os.Unsetenv("AETERNA_INHERITED_FDS")
	sm := NewSocketManager()

	// 1. Initial creation
	l1, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("First EnsureListener failed: %v", err)
	}

	// 2. Close it
	sm.Close()

	if sm.GetFile() != nil {
		t.Fatal("File should be nil after Close()")
	}

	// 3. Try to get it again
	l2, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Second EnsureListener failed: %v", err)
	}

	if l1 == l2 {
		t.Fatal("Expected a new listener instance after Close(), but got the same one")
	}

	// Verify the new listener is usable
	if l2.Addr() == nil {
		t.Fatal("New listener should have a valid address")
	}
	sm.Close()
}

// Note: Testing inheritance requires spawning a real subprocess which is complex for unit test.
// We mock the Env check logic in a real integration test.

// Personal.AI order the ending
