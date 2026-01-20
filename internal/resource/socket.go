package resource

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"syscall"

	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/logger"
)

type SocketManager struct {
	mu sync.Mutex

	// Active listeners keyed by address
	listeners map[string]net.Listener
	files     map[string]*os.File

	// Inherited but not yet claimed listeners
	inherited map[string]*inheritedSocket

	discovered bool
}

type inheritedSocket struct {
	listener net.Listener
	file     *os.File
}

func NewSocketManager() *SocketManager {
	return &SocketManager{
		listeners: make(map[string]net.Listener),
		files:     make(map[string]*os.File),
		inherited: make(map[string]*inheritedSocket),
	}
}

func isSocket(fd uintptr) bool {
	var stat syscall.Stat_t
	err := syscall.Fstat(int(fd), &stat)
	if err != nil {
		return false
	}
	return (stat.Mode & syscall.S_IFMT) == syscall.S_IFSOCK
}

func (sm *SocketManager) discoverInherited() {
	if sm.discovered {
		return
	}
	sm.discovered = true

	fds := os.Getenv(consts.EnvInheritedFDs)
	if fds == "" {
		return
	}

	count, err := strconv.Atoi(fds)
	if err != nil || count <= 0 {
		return
	}

	// Clear it so children of this process don't see it unless we set it again
	os.Unsetenv(consts.EnvInheritedFDs)

	logger.Log.Info("Hot Relay: Discovering inherited sockets", "count", count)

	for i := 0; i < count; i++ {
		// ExtraFiles start at fd 3
		fd := 3 + i
		if !isSocket(uintptr(fd)) {
			logger.Log.Warn("Hot Relay: FD is not a socket, skipping", "fd", fd)
			continue
		}

		f := os.NewFile(uintptr(fd), "listener")
		if f == nil {
			continue
		}

		l, err := net.FileListener(f)
		if err != nil {
			logger.Log.Error("Hot Relay: Failed to create listener from FD", "fd", fd, "err", err)
			// We don't close f here because if it failed, we might not truly "own" this FD
			// especially in test environments.
			continue
		}

		// Ensure non-blocking mode for Go runtime poller
		if tcpL, ok := l.(*net.TCPListener); ok {
			if rawConn, err := tcpL.SyscallConn(); err == nil {
				rawConn.Control(func(fd uintptr) {
					_ = syscall.SetNonblock(int(fd), true)
				})
			}
		}

		addr := l.Addr().String()
		sm.inherited[addr] = &inheritedSocket{
			listener: l,
			file:     f,
		}
		logger.Log.Info("Hot Relay: Discovered inherited socket", "addr", addr, "fd", fd)
	}
}

// EnsureListener returns a listener, either inherited from parent or created new
func (sm *SocketManager) EnsureListener(addr string) (net.Listener, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Check if we already have it active
	if l, ok := sm.listeners[addr]; ok {
		return l, nil
	}

	// 2. Try to discover inherited sockets if not already done
	sm.discoverInherited()

	// 3. Check if it was inherited
	if is, ok := sm.inherited[addr]; ok {
		logger.Log.Info("Hot Relay: Claiming inherited socket", "addr", addr)
		sm.listeners[addr] = is.listener
		sm.files[addr] = is.file
		delete(sm.inherited, addr)
		return is.listener, nil
	}

	// 4. Cold start
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

	// File() sets the socket to blocking mode. We need to set it back to non-blocking
	if rawConn, err := tcpL.SyscallConn(); err == nil {
		rawConn.Control(func(fd uintptr) {
			_ = syscall.SetNonblock(int(fd), true)
		})
	}

	sm.listeners[addr] = l
	sm.files[addr] = f
	return l, nil
}

// GetFiles returns all managed file descriptors to pass to child.
// The files are returned in a deterministic order (sorted by address).
func (sm *SocketManager) GetFiles() []*os.File {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Get addresses and sort them for determinism
	addrs := make([]string, 0, len(sm.files))
	for addr := range sm.files {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)

	files := make([]*os.File, 0, len(sm.files))
	for _, addr := range addrs {
		files = append(files, sm.files[addr])
	}
	return files
}

// GetFile is deprecated, use GetFiles
func (sm *SocketManager) GetFile() *os.File {
	files := sm.GetFiles()
	if len(files) > 0 {
		return files[0]
	}
	return nil
}

func (sm *SocketManager) Close() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for addr, l := range sm.listeners {
		l.Close()
		if f, ok := sm.files[addr]; ok {
			f.Close()
		}
	}
	sm.listeners = make(map[string]net.Listener)
	sm.files = make(map[string]*os.File)

	for _, is := range sm.inherited {
		is.listener.Close()
		is.file.Close()
	}
	sm.inherited = make(map[string]*inheritedSocket)
}

// Personal.AI order the ending
