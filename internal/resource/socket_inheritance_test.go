package resource

import (
	"os"
	"testing"
	"github.com/turtacn/Aeterna/pkg/consts"
)

func TestEnsureListener_InheritanceBugFixed(t *testing.T) {
	// Set the env var to something that will cause inheritance to be attempted.
    // FD 3 might not even exist or might not be a socket.
	os.Setenv(consts.EnvInheritedFDs, "1")
	defer os.Unsetenv(consts.EnvInheritedFDs)

	sm := NewSocketManager()

	// First call - should attempt to inherit FD 3.
    // Whether it succeeds or fails to inherit, it should clear the env var.
    // If it fails to inherit, it should fallback to cold start and succeed.
	l1, err := sm.EnsureListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("EnsureListener failed: %v", err)
	}
	if l1 == nil {
		t.Fatal("Expected l1 to be non-nil")
	}
    defer l1.Close()

    // Verify env var is cleared after use
    if os.Getenv(consts.EnvInheritedFDs) != "" {
        t.Error("AETERNA_INHERITED_FDS should be cleared after being processed")
    }

    // Call again - should definitely do a cold start (which it does anyway if addr is same)
    // but the point is it shouldn't try to inherit again.
    l2, err := sm.EnsureListener("127.0.0.1:0")
    if err != nil {
        t.Fatalf("Second call failed: %v", err)
    }
    l2.Close()
}
