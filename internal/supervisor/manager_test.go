package supervisor

import (
	"os"
	"testing"
)

func TestProcessManager_StartStop(t *testing.T) {
	pm := New()

	// Start a simple long-running command
	err := pm.Start([]string{"sleep", "10"}, nil, nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if pm.cmd.Process == nil {
		t.Fatal("Process should be started")
	}

	err = pm.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// Give it a moment to exit
	err = pm.Wait()
	if err == nil {
		// sleep 10 might return nil if it handles SIGTERM gracefully or 0
		// but usually it's interrupted
	}
}

func TestProcessManager_Kill(t *testing.T) {
	pm := New()
	err := pm.Start([]string{"sleep", "10"}, nil, nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = pm.Kill()
	if err != nil {
		t.Errorf("Kill failed: %v", err)
	}

	err = pm.Wait()
	if err == nil {
		t.Errorf("Wait should have returned error for killed process")
	}
}

func TestProcessManager_EmptyCommand(t *testing.T) {
	pm := New()
	err := pm.Start(nil, nil, nil)
	if err != nil {
		t.Errorf("Start(nil) should not return error, got %v", err)
	}
}

func TestProcessManager_WithFiles(t *testing.T) {
	pm := New()
	f, _ := os.Open(os.DevNull)
	defer f.Close()

	err := pm.Start([]string{"ls"}, nil, []*os.File{f})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	pm.Wait()
}
