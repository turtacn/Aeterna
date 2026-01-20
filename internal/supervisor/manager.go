package supervisor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/logger"
)

// ProcessManager handles the lifecycle of the managed business process.
// It manages starting, stopping, and waiting for the process.
type ProcessManager struct {
	cmd *exec.Cmd
}

// New creates a new ProcessManager instance.
func New() *ProcessManager {
	return &ProcessManager{}
}

// Start launches the business process with the given command, environment, and extra files.
// It sets up standard output and error redirection and communicates inherited
// file descriptors to the child process via the AETERNA_INHERITED_FDS environment variable.
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

// Stop sends a SIGTERM signal to the managed process to initiate a graceful shutdown.
func (pm *ProcessManager) Stop() error {
	if pm.cmd != nil && pm.cmd.Process != nil {
		logger.Log.Info("Supervisor: Sending SIGTERM", "pid", pm.cmd.Process.Pid)
		return pm.cmd.Process.Signal(syscall.SIGTERM)
	}
	return nil
}

// Kill immediately terminates the managed process using a SIGKILL signal.
// This is typically used during rollbacks if a graceful shutdown fails.
func (pm *ProcessManager) Kill() error {
	if pm.cmd != nil && pm.cmd.Process != nil {
		logger.Log.Warn("Supervisor: Sending SIGKILL (Rollback)", "pid", pm.cmd.Process.Pid)
		return pm.cmd.Process.Kill()
	}
	return nil
}

// Wait waits for the managed process to exit and returns the resulting error, if any.
func (pm *ProcessManager) Wait() error {
	if pm.cmd != nil {
		return pm.cmd.Wait()
	}
	return nil
}

// Personal.AI order the ending
