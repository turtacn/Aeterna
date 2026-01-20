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
	baseFD     int
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
		baseFD:    3,
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
		// ExtraFiles start at baseFD (usually 3)
		fd := sm.baseFD + i
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

func (sm *SocketManager) addressesMatch(a, b string) bool {
	if a == b {
		return true
	}
	ra, err1 := net.ResolveTCPAddr("tcp", a)
	rb, err2 := net.ResolveTCPAddr("tcp", b)
	if err1 != nil || err2 != nil {
		return false
	}
	if ra.Port != rb.Port {
		return false
	}

	isWildcardA := ra.IP == nil || ra.IP.IsUnspecified()
	isWildcardB := rb.IP == nil || rb.IP.IsUnspecified()

	if isWildcardA && isWildcardB {
		return true
	}

	return ra.IP.Equal(rb.IP)
}

func (sm *SocketManager) findListenerLocked(addr string) (net.Listener, *os.File) {
	if l, ok := sm.listeners[addr]; ok {
		return l, sm.files[addr]
	}

	// For random port (0), we never match an existing listener unless it's by exact string key (which is unlikely for :0)
	if _, port, err := net.SplitHostPort(addr); err == nil && port == "0" {
		return nil, nil
	}

	for can, l := range sm.listeners {
		if sm.addressesMatch(addr, can) {
			return l, sm.files[can]
		}
	}
	return nil, nil
}

func (sm *SocketManager) findInheritedLocked(addr string) *inheritedSocket {
	if is, ok := sm.inherited[addr]; ok {
		return is
	}

	// For random port (0), we can't match an inherited socket by address unless it's exact match
	if _, port, err := net.SplitHostPort(addr); err == nil && port == "0" {
		return nil
	}

	for can, is := range sm.inherited {
		if sm.addressesMatch(addr, can) {
			return is
		}
	}
	return nil
}

// EnsureListener returns a listener, either inherited from parent or created new
func (sm *SocketManager) EnsureListener(addr string) (net.Listener, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Check if we already have it active
	if l, f := sm.findListenerLocked(addr); l != nil {
		// Map this alias for future lookups and inheritance
		sm.listeners[addr] = l
		if f != nil {
			sm.files[addr] = f
		}
		return l, nil
	}

	// 2. Try to discover inherited sockets if not already done
	sm.discoverInherited()

	// 3. Check if it was inherited
	if is := sm.findInheritedLocked(addr); is != nil {
		canonicalAddr := is.listener.Addr().String()
		logger.Log.Info("Hot Relay: Claiming inherited socket", "requested", addr, "canonical", canonicalAddr)
		sm.listeners[addr] = is.listener
		sm.files[addr] = is.file
		if canonicalAddr != addr {
			sm.listeners[canonicalAddr] = is.listener
			sm.files[canonicalAddr] = is.file
		}
		delete(sm.inherited, canonicalAddr)
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
	canonicalAddr := l.Addr().String()
	if canonicalAddr != addr {
		sm.listeners[canonicalAddr] = l
		sm.files[canonicalAddr] = f
	}
	return l, nil
}

// GetFiles returns all managed file descriptors to pass to child.
// The files are returned in a deterministic order (sorted by address).
func (sm *SocketManager) GetFiles() []*os.File {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Ensure we've discovered all inherited sockets so they can be passed down
	sm.discoverInherited()

	// Use a map to deduplicate files since a listener might be stored under multiple keys
	uniqueFiles := make(map[uintptr]*os.File)
	addrMap := make(map[uintptr]string)

	// Include both active files and inherited but unclaimed files
	allSources := make(map[string]*os.File)
	for addr, f := range sm.files {
		allSources[addr] = f
	}
	for addr, is := range sm.inherited {
		allSources[addr] = is.file
	}

	for addr, f := range allSources {
		fd := f.Fd()
		if _, ok := uniqueFiles[fd]; !ok {
			uniqueFiles[fd] = f
			addrMap[fd] = addr
		} else {
			// If we have multiple addresses for the same FD, prefer the canonical one (longer usually)
			// or just stay consistent. Canonical addresses are often longer e.g. [::]:8080 vs :8080
			if len(addr) > len(addrMap[fd]) {
				addrMap[fd] = addr
			}
		}
	}

	// Sort by address for determinism
	type fileWithAddr struct {
		f    *os.File
		addr string
	}
	sorted := make([]fileWithAddr, 0, len(uniqueFiles))
	for fd, f := range uniqueFiles {
		sorted = append(sorted, fileWithAddr{f, addrMap[fd]})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].addr < sorted[j].addr
	})

	files := make([]*os.File, 0, len(sorted))
	for _, s := range sorted {
		files = append(files, s.f)
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
