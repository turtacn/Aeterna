package srp

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateCoordinator_PrepareSocket(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")
	sc := NewCoordinator(socketPath)

	l, err := sc.PrepareSocket()
	if err != nil {
		t.Fatalf("PrepareSocket failed: %v", err)
	}
	defer l.Close()

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Errorf("Socket file %s should exist", socketPath)
	}

	// Test cleanup and recreation
	l2, err := sc.PrepareSocket()
	if err != nil {
		t.Fatalf("PrepareSocket failed on second call: %v", err)
	}
	l2.Close()
}

func TestStateCoordinator_WaitStateTransfer_Success(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")
	sc := NewCoordinator(socketPath)

	testData := map[string]interface{}{
		"key1": "value1",
		"key2": float64(42),
	}

	// Start WaitStateTransfer in a goroutine
	done := make(chan bool)
	go func() {
		state, err := sc.WaitStateTransfer(2 * time.Second)
		if err != nil {
			t.Errorf("WaitStateTransfer failed: %v", err)
		}
		if state["key1"] != testData["key1"] || state["key2"] != testData["key2"] {
			t.Errorf("Received state mismatch: got %v, want %v", state, testData)
		}
		done <- true
	}()

	// Give it a moment to start listening
	time.Sleep(100 * time.Millisecond)

	// Simulate old process sending state
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(testData); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestStateCoordinator_WaitStateTransfer_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test_timeout.sock")
	sc := NewCoordinator(socketPath)

	_, err := sc.WaitStateTransfer(100 * time.Millisecond)
	if err != os.ErrDeadlineExceeded {
		t.Errorf("Expected deadline exceeded error, got %v", err)
	}
}

func TestStateCoordinator_WaitStateTransfer_InvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test_invalid.sock")
	sc := NewCoordinator(socketPath)

	// Start WaitStateTransfer in a goroutine
	done := make(chan bool)
	go func() {
		_, err := sc.WaitStateTransfer(2 * time.Second)
		if err == nil {
			t.Errorf("WaitStateTransfer should have failed with invalid data")
		}
		done <- true
	}()

	// Give it a moment to start listening
	time.Sleep(100 * time.Millisecond)

	// Simulate old process sending invalid data
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	conn.Write([]byte("invalid json"))
	conn.Close() // Close to trigger EOF or decode error

	select {
	case <-done:
		// success (it failed as expected)
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out")
	}
}
