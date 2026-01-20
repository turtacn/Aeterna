package srp

import (
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/turtacn/Aeterna/pkg/logger"
)

// StateCoordinator manages the State Relay Protocol (SRP) process.
// It uses a Unix domain socket to facilitate memory context transfer between
// an old process and a new process during a hot reload.
type StateCoordinator struct {
	socketPath string
}

// NewCoordinator creates a new StateCoordinator with the specified socket path.
func NewCoordinator(path string) *StateCoordinator {
	return &StateCoordinator{socketPath: path}
}

// PrepareSocket creates and returns a Unix domain socket listener for state transfer.
// It removes any existing socket at the path and sets appropriate file permissions.
func (sc *StateCoordinator) PrepareSocket() (net.Listener, error) {
	if _, err := os.Stat(sc.socketPath); err == nil {
		os.Remove(sc.socketPath)
	}
	l, err := net.Listen("unix", sc.socketPath)
	if err != nil {
		return nil, err
	}
	// Ensure permissions allow the process to read/write
	os.Chmod(sc.socketPath, 0700)
	return l, nil
}

// WaitStateTransfer waits for the old process to dump its state via the Unix socket.
// It returns the decoded state data or an error if the transfer fails or times out.
// This is typically called by the new process during its startup phase.
func (sc *StateCoordinator) WaitStateTransfer(timeout time.Duration) (map[string]interface{}, error) {
	logger.Log.Info("SRP: Waiting for state handover...", "socket", sc.socketPath)

	l, err := sc.PrepareSocket()
	if err != nil {
		return nil, err
	}
	defer l.Close()
	defer os.Remove(sc.socketPath)

	type result struct {
		data map[string]interface{}
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		conn, err := l.Accept()
		if err != nil {
			ch <- result{nil, err}
			return
		}
		defer conn.Close()

		conn.SetReadDeadline(time.Now().Add(timeout))

		var state map[string]interface{}
		// In production, use Protobuf for efficiency. Here using JSON for readability.
		decoder := json.NewDecoder(conn)
		if err := decoder.Decode(&state); err != nil {
			ch <- result{nil, err}
			return
		}

		logger.Log.Info("SRP: Context received", "keys", len(state))
		ch <- result{state, nil}
	}()

	select {
	case res := <-ch:
		return res.data, res.err
	case <-time.After(timeout):
		return nil, os.ErrDeadlineExceeded
	}
}

// Personal.AI order the ending
