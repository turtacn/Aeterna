package resource

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/logger"
)

type SocketManager struct {
	mu          sync.Mutex
	listener    net.Listener
	file        *os.File
	currentAddr string
}

func NewSocketManager() *SocketManager {
	return &SocketManager{}
}

// EnsureListener returns a listener, either inherited from parent or created new
func (sm *SocketManager) EnsureListener(addr string) (net.Listener, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.listener != nil {
		if sm.currentAddr == addr {
			return sm.listener, nil
		}
		// Address changed, close existing
		sm.listener.Close()
		if sm.file != nil {
			sm.file.Close()
		}
		sm.listener = nil
		sm.file = nil
		sm.currentAddr = ""
	}

	// 1. Check if we are running as a child process with inherited FDs
	fds := os.Getenv(consts.EnvInheritedFDs)
	if fds != "" {
		// Clear it so we don't try to inherit again if this fails or is called again
		os.Unsetenv(consts.EnvInheritedFDs)

		count, err := strconv.Atoi(fds)
		if err == nil && count > 0 {
			logger.Log.Info("Hot Relay: Inheriting socket from parent", "fds", count)
			// ExtraFiles start at fd 3 (0:stdin, 1:stdout, 2:stderr)
			f := os.NewFile(3, "listener")
			l, err := net.FileListener(f)
			if err != nil {
				// Fallback to cold start if inheritance fails
				logger.Log.Error("Hot Relay: Failed to inherit socket, falling back to cold start", "err", err)
			} else {
				sm.file = f
				sm.listener = l
				sm.currentAddr = addr
				return l, nil
			}
		}
	}

	// 2. Cold start
	logger.Log.Info("Cold Start: Binding new listener", "addr", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Get the file descriptor for future inheritance
	tcpL, ok := l.(*net.TCPListener)
	if !ok {
		l.Close()
		return nil, fmt.Errorf("listener is not a TCP listener")
	}
	f, err := tcpL.File()
	if err != nil {
		l.Close()
		return nil, err
	}

	sm.listener = l
	sm.file = f
	sm.currentAddr = addr
	return l, nil
}

// GetFile returns the file descriptor to pass to child
func (sm *SocketManager) GetFile() *os.File {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.file
}

func (sm *SocketManager) Close() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.listener != nil {
		sm.listener.Close()
		sm.listener = nil
	}
	if sm.file != nil {
		sm.file.Close()
		sm.file = nil
	}
	sm.currentAddr = ""
}

// Personal.AI order the ending
