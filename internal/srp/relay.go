package srp

import (
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/turtacn/Aeterna/pkg/logger"
)

type StateCoordinator struct {
	socketPath string
}

func NewCoordinator(path string) *StateCoordinator {
	return &StateCoordinator{socketPath: path}
}

// PrepareSocket creates a Unix domain socket for the child to connect to
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

// WaitStateTransfer waits for the OLD process to dump its state
// This function is called by the NEW process (Orchestrator) during startup.
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
