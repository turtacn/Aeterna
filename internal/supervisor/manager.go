package supervisor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/logger"
)

type ProcessManager struct {
	cmd *exec.Cmd
}

func New() *ProcessManager {
	return &ProcessManager{}
}

// Start launches the business process
func (pm *ProcessManager) Start(command []string, env []string, extraFiles []*os.File) error {
	if len(command) == 0 {
		return nil
	}

	pm.cmd = exec.Command(command[0], command[1:]...)
	pm.cmd.Env = append(os.Environ(), env...)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr

	if len(extraFiles) > 0 {
		pm.cmd.ExtraFiles = extraFiles
		// UPHR Core: Notify child about inherited FDs
		pm.cmd.Env = append(pm.cmd.Env, fmt.Sprintf("%s=%d", consts.EnvInheritedFDs, len(extraFiles)))
	}

	logger.Log.Info("Supervisor: Forking process", "cmd", command)
	return pm.cmd.Start()
}

// Stop gracefully terminates the process
func (pm *ProcessManager) Stop() error {
	if pm.cmd != nil && pm.cmd.Process != nil {
		logger.Log.Info("Supervisor: Sending SIGTERM", "pid", pm.cmd.Process.Pid)
		return pm.cmd.Process.Signal(syscall.SIGTERM)
	}
	return nil
}

// Kill immediately terminates the process (Used for Rollback)
func (pm *ProcessManager) Kill() error {
	if pm.cmd != nil && pm.cmd.Process != nil {
		logger.Log.Warn("Supervisor: Sending SIGKILL (Rollback)", "pid", pm.cmd.Process.Pid)
		return pm.cmd.Process.Kill()
	}
	return nil
}

func (pm *ProcessManager) Wait() error {
	if pm.cmd != nil {
		return pm.cmd.Wait()
	}
	return nil
}

// Personal.AI order the ending
