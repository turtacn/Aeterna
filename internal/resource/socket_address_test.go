package resource

import (
	"os"
	"testing"
)

func TestSocketManager_AddressChange(t *testing.T) {
	os.Unsetenv("AETERNA_INHERITED_FDS")
	sm := NewSocketManager()
	defer sm.Close()

	// First call to bind to a random port
	l1, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call with a different address (127.0.0.2:0)
	l2, err := sm.EnsureListener("127.0.0.2:0")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if l1 == l2 {
		t.Errorf("Expected different listener instance when address changes, but got the same one")
	}
}
