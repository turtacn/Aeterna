package consts

import "time"

// AppMode defines the running mode of Aeterna.
type AppMode string

const (
	ModeRun      AppMode = "run"       // Normal supervisor mode
	ModeHotRelay AppMode = "hot-relay" // Mode allowing in-place updates
)

// ProcessState defines the detailed lifecycle state of the managed process.
// It is based on the UPHR-O (Universal Process Hot-Relay Orchestration) Architecture.
type ProcessState string

const (
	StatePending     ProcessState = "PENDING"
	StatePreChecking ProcessState = "PRE_CHECKING" // Phase 1: Dry-run/Config check
	StateStarting    ProcessState = "STARTING"     // Phase 2: Fork & Exec
	StateHandshaking ProcessState = "HANDSHAKING"  // Phase 2.5: SRP Context Transfer
	StateSoaking     ProcessState = "SOAKING"      // Phase 3: Canary/Soak
	StateRunning     ProcessState = "RUNNING"      // Stable
	StateDraining    ProcessState = "DRAINING"     // Phase 5: Old process shutdown
	StateStopped     ProcessState = "STOPPED"
	StateFailed      ProcessState = "FAILED"
)

// SRP (State Relay Protocol) Constants
const (
	EnvStateSocketPath = "AETERNA_STATE_SOCK"
	EnvInheritedFDs    = "AETERNA_INHERITED_FDS" // Count of FDs passed
	DefaultSRPTimeout  = 5 * time.Second
	DefaultSoakTime    = 30 * time.Second
)

// Personal.AI order the ending
