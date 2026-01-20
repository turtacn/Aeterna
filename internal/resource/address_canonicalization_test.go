package resource

import (
	"testing"
	"strconv"
    "net"
    "os"
    "github.com/turtacn/Aeterna/pkg/consts"
    "syscall"
)

func TestSocketManager_AddressNormalization(t *testing.T) {
	sm := NewSocketManager()
	defer sm.Close()

	port := 45005
	addr1 := ":" + strconv.Itoa(port)

	l1, err := sm.EnsureListener(addr1)
	if err != nil {
		t.Fatalf("Failed to create first listener: %v", err)
	}

    canonicalAddr := l1.Addr().String()
    t.Logf("Requested %s, got %s", addr1, canonicalAddr)

    if addr1 == canonicalAddr {
        t.Skip("Requested address is already canonical on this system")
    }

	l2, err := sm.EnsureListener(canonicalAddr)
	if err != nil {
		t.Fatalf("Failed to get listener with canonical address: %v. This is likely because it tried to bind again and failed with 'address already in use'", err)
	}

	if l1 != l2 {
		t.Errorf("Expected same listener instance for %s and %s", addr1, canonicalAddr)
	}
}

func TestSocketManager_InheritanceAddressNormalization(t *testing.T) {
	// 1. Create a listener and get its canonical address
	l, err := net.Listen("tcp", ":0") // Bind to all interfaces
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	canonicalAddr := l.Addr().String()
    _, portStr, _ := net.SplitHostPort(canonicalAddr)
    requestedAddr := ":" + portStr

    if requestedAddr == canonicalAddr {
         l.Close()
         t.Skip("Requested address is already canonical on this system, cannot demonstrate mismatch")
    }

	tcpL := l.(*net.TCPListener)
	f, _ := tcpL.File()
	l.Close()

	// 2. Set up environment for inheritance
	os.Setenv(consts.EnvInheritedFDs, "1")
	defer os.Unsetenv(consts.EnvInheritedFDs)

	// Mock FD 3
	err = syscall.Dup2(int(f.Fd()), 3)
	if err != nil {
		t.Fatalf("Failed to dup2: %v", err)
	}
	defer syscall.Close(3)
    f.Close()

	sm := NewSocketManager()
	defer sm.Close()

	// 3. Try to claim it using the requested address string
    t.Logf("Inherited %s, trying to claim with %s", canonicalAddr, requestedAddr)
	l2, err := sm.EnsureListener(requestedAddr)
	if err != nil {
		t.Fatalf("Failed to claim inherited listener with requested address: %v", err)
	}

	if l2.Addr().String() != canonicalAddr {
		t.Errorf("Expected address %s, got %s", canonicalAddr, l2.Addr().String())
	}

    // Check if it was actually inherited or if it was a cold start (it should fail if cold start)
    // Actually, if it was a cold start, it would fail because the port is still held by FD 3 (or the dup in SocketManager)
}
