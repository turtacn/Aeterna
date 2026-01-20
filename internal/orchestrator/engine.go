package orchestrator

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/Aeterna/internal/resource"
	"github.com/turtacn/Aeterna/internal/srp"
	"github.com/turtacn/Aeterna/internal/supervisor"
	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/fsm"
	"github.com/turtacn/Aeterna/pkg/logger"
	"github.com/turtacn/Aeterna/pkg/protocol"
)

type Engine struct {
	cfg     *protocol.Config
	fsm     *fsm.StateMachine
	socket  *resource.SocketManager
	process *supervisor.ProcessManager // Current Process
	srp     *srp.StateCoordinator
}

func NewEngine(cfg *protocol.Config) *Engine {
	e := &Engine{
		cfg:     cfg,
		fsm:     fsm.New(fsm.State(consts.StatePending)),
		socket:  resource.NewSocketManager(),
		process: supervisor.New(),
		srp:     srp.NewCoordinator(cfg.Orchestration.StateHandoff.SocketPath),
	}
	e.setupFSM()
	return e
}

func (e *Engine) setupFSM() {
	// Define UPHR-O State Transitions

	// Initial Start
	e.fsm.AddTransition(fsm.State(consts.StatePending), fsm.State(consts.StateStarting), "start", e.onStart)
	e.fsm.AddTransition(fsm.State(consts.StateStarting), fsm.State(consts.StateRunning), "stable", nil)

	// Hot Reload Trigger
	e.fsm.AddTransition(fsm.State(consts.StateRunning), fsm.State(consts.StatePreChecking), "reload", e.onReloadTriggered)

	// Reload Flow
	e.fsm.AddTransition(fsm.State(consts.StatePreChecking), fsm.State(consts.StateRunning), "abort", nil) // Check failed
	e.fsm.AddTransition(fsm.State(consts.StatePreChecking), fsm.State(consts.StateSoaking), "proceed", e.onSoakStart)

	// Soak Outcome
	e.fsm.AddTransition(fsm.State(consts.StateSoaking), fsm.State(consts.StateRunning), "rollback", e.onRollback)
	e.fsm.AddTransition(fsm.State(consts.StateSoaking), fsm.State(consts.StateDraining), "success", e.onDrainOld)
}

func (e *Engine) Start() error {
	// Handle OS Signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigCh {
			switch sig {
			case syscall.SIGHUP:
				logger.Log.Info("Signal: SIGHUP received. Initiating UPHR-O workflow.")
				e.fsm.Fire("reload")
			case syscall.SIGINT, syscall.SIGTERM:
				logger.Log.Info("Signal: Stop received. Shutting down.")
				e.process.Stop()
				os.Exit(0)
			}
		}
	}()

	// Initial bootstrap
	return e.fsm.Fire("start")
}

// onStart handles the initial cold start
func (e *Engine) onStart(event fsm.Event, args ...interface{}) error {
	logger.Log.Info("Phase: Cold Start")

	// 1. Bind Socket
	_, err := e.socket.EnsureListener(":8080") // In real code, parse from config
	if err != nil {
		return err
	}

	// 2. Start Process
	files := e.socket.GetFiles()

	err = e.process.Start(e.cfg.Service.Command, e.cfg.Service.Env, files)
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(2 * time.Second) // Simple warmup
		e.fsm.Fire("stable")
	}()

	// Block and wait for process
	return e.process.Wait()
}

// onReloadTriggered: Phase 1 - Pre-flight Checks
func (e *Engine) onReloadTriggered(event fsm.Event, args ...interface{}) error {
	logger.Log.Info("Phase 1: Pre-flight Checks")

	for _, hook := range e.cfg.Orchestration.PreFlight {
		logger.Log.Info("Running hook", "name", hook.Name)
		cmd := exec.Command(hook.Command[0], hook.Command[1:]...)
		if err := cmd.Run(); err != nil {
			logger.Log.Error("Pre-flight check failed. Aborting reload.", "hook", hook.Name, "err", err)
			e.fsm.Fire("abort")
			return nil // Return nil to state machine, effectively handled
		}
	}

	logger.Log.Info("Pre-flight checks passed.")
	e.fsm.Fire("proceed")
	return nil
}

// onSoakStart: Phase 2 & 3 - Fork, Exec & Soak
func (e *Engine) onSoakStart(event fsm.Event, args ...interface{}) error {
	logger.Log.Info("Phase 2 & 3: Forking New Process & Soaking")

	// Note: In a real implementation, we need to manage Two ProcessManagers (old and new).
	// For this blueprint, we simulate the decision logic.

	soakDuration, _ := time.ParseDuration(e.cfg.Orchestration.Canary.SoakTime)
	if soakDuration == 0 {
		soakDuration = consts.DefaultSoakTime
	}

	go func() {
		logger.Log.Info("Soaking...", "duration", soakDuration)
		// Monitoring simulation
		time.Sleep(soakDuration)

		// If metrics are good:
		success := true

		if success {
			e.fsm.Fire("success")
		} else {
			e.fsm.Fire("rollback")
		}
	}()

	return nil
}

func (e *Engine) onRollback(event fsm.Event, args ...interface{}) error {
	logger.Log.Warn("Phase: Rollback. Killing new process.")
	// Logic to kill new process
	return nil
}

func (e *Engine) onDrainOld(event fsm.Event, args ...interface{}) error {
	logger.Log.Info("Phase 5: Drain. Stopping old process.")
	// Logic to signal old process to exit
	// Trigger Post-processing hooks
	return nil
}

// Personal.AI order the ending
