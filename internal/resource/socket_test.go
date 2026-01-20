package resource

import (
	"testing"
)

func TestSocketManager_MultipleListeners(t *testing.T) {
	sm := NewSocketManager()
	defer sm.Close()

	// Use specific ports to avoid collisions and 0-port ambiguity
	l1, err := sm.EnsureListener("127.0.0.1:45001")
	if err != nil {
		t.Fatalf("Failed to create first listener: %v", err)
	}
	addr1 := l1.Addr().String()

	l2, err := sm.EnsureListener("127.0.0.1:45002")
	if err != nil {
		t.Fatalf("Failed to create second listener: %v", err)
	}
	addr2 := l2.Addr().String()

	if addr1 == addr2 {
		t.Errorf("Expected different addresses, got both as %s.", addr1)
	}

	// Verify both are still usable and return same instance
	l1Again, err := sm.EnsureListener(addr1)
	if err != nil {
		t.Fatalf("Failed to get first listener again: %v", err)
	}
	if l1Again != l1 {
		t.Errorf("Expected same listener instance for %s", addr1)
	}

	l2Again, err := sm.EnsureListener(addr2)
	if err != nil {
		t.Fatalf("Failed to get second listener again: %v", err)
	}
	if l2Again != l2 {
		t.Errorf("Expected same listener instance for %s", addr2)
	}
}
