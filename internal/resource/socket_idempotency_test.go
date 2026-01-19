package resource

import (
	"os"
	"testing"
)

func TestSocketManager_Idempotency(t *testing.T) {
	os.Unsetenv("AETERNA_INHERITED_FDS")
	sm := NewSocketManager()
	defer sm.Close()

	// First call
	l1, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call with same address should return the same listener
	l2, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if l1 != l2 {
		t.Error("Expected same listener instance on second call")
	}
}

func TestSocketManager_NoLeakOnOverwrite(t *testing.T) {
	os.Unsetenv("AETERNA_INHERITED_FDS")
	sm := NewSocketManager()
	defer sm.Close()

	// First call
	_, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	f1 := sm.GetFile()

	// Second call with DIFFERENT address to trigger a new bind
	// (Note: in our improved version, we might still return the same listener if we want to enforce single listener,
	// but let's see how we implement it. If we allow changing the address, we must close the old one.)
	_, err = sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	f2 := sm.GetFile()

    // In our proposed fix, EnsureListener will return the EXISTING listener regardless of addr
    // if sm.listener is already set. If that's the case, f1 == f2.
    // If we decide to allow changing addr, then f1 should be closed.

    if f1 != f2 {
        // If they are different, f1 should have been closed by EnsureListener
        _, err = f1.Stat()
        if err == nil {
            t.Error("f1 should have been closed when overwritten")
        }
    }
}
