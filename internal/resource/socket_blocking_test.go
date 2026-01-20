package resource

import (
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/turtacn/Aeterna/pkg/consts"
	"golang.org/x/sys/unix"
)

func isNonblocking(fd uintptr) (bool, error) {
	flags, err := unix.FcntlInt(fd, unix.F_GETFL, 0)
	if err != nil {
		return false, err
	}
	return flags&unix.O_NONBLOCK != 0, nil
}

func TestInheritedSocketNonBlocking(t *testing.T) {
	// 1. Create a listener and get its file descriptor
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	tcpL := l.(*net.TCPListener)
	f, err := tcpL.File()
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}
	defer f.Close()

	// 2. Mock inheritance by setting up FD 3
	// Save original FD 3 if it exists
	origFd3, err := syscall.Dup(3)
	hasOrig := err == nil
	if hasOrig {
		defer func() {
			syscall.Dup2(origFd3, 3)
			syscall.Close(origFd3)
		}()
	}

	err = syscall.Dup2(int(f.Fd()), 3)
	if err != nil {
		t.Fatalf("Failed to dup2: %v", err)
	}
	defer syscall.Close(3)

	// Set environment variable
	os.Setenv(consts.EnvInheritedFDs, "1")
	defer os.Unsetenv(consts.EnvInheritedFDs)

	// 3. Use SocketManager to inherit
	sm := NewSocketManager()
	inheritedL, err := sm.EnsureListener(l.Addr().String())
	if err != nil {
		t.Fatalf("EnsureListener failed: %v", err)
	}
	defer sm.Close()

	// 4. Check if the inherited listener is non-blocking
	inheritedTCP, ok := inheritedL.(*net.TCPListener)
	if !ok {
		t.Fatal("Inherited listener is not a TCP listener")
	}

	rawConn, err := inheritedTCP.SyscallConn()
	if err != nil {
		t.Fatalf("Failed to get SyscallConn: %v", err)
	}

	var nonblocking bool
	var controlErr error
	err = rawConn.Control(func(fd uintptr) {
		nonblocking, controlErr = isNonblocking(fd)
	})
	if err != nil {
		t.Fatalf("Control failed: %v", err)
	}
	if controlErr != nil {
		t.Fatalf("isNonblocking failed: %v", controlErr)
	}

	if !nonblocking {
		t.Error("Inherited listener is in blocking mode, expected non-blocking")
	}
}
