package resource

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/logger"
)

type SocketManager struct {
	listener net.Listener
	file     *os.File
}

func NewSocketManager() *SocketManager {
	return &SocketManager{}
}

// EnsureListener returns a listener, either inherited from parent or created new
func (sm *SocketManager) EnsureListener(addr string) (net.Listener, error) {
	if sm.listener != nil {
		return sm.listener, nil
	}

	// 1. Check if we are running as a child process with inherited FDs
	fds := os.Getenv(consts.EnvInheritedFDs)
	if fds != "" {
		count, err := strconv.Atoi(fds)
		if err == nil && count > 0 {
			logger.Log.Info("Hot Relay: Inheriting socket from parent", "fds", count)
			// ExtraFiles start at fd 3 (0:stdin, 1:stdout, 2:stderr)
			f := os.NewFile(3, "listener")
			l, err := net.FileListener(f)
			if err != nil {
				return nil, fmt.Errorf("failed to recreate listener from fd: %w", err)
			}
			sm.file = f
			sm.listener = l
			return l, nil
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
	return l, nil
}

// GetFile returns the file descriptor to pass to child
func (sm *SocketManager) GetFile() *os.File {
	return sm.file
}

func (sm *SocketManager) Close() {
	if sm.listener != nil {
		sm.listener.Close()
	}
	if sm.file != nil {
		sm.file.Close()
	}
}

// Personal.AI order the ending
