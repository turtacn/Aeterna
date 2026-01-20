package protocol

// Config represents the root configuration compliant with UPHR-O.
type Config struct {
	Version       string              `yaml:"version"`
	Service       ServiceConfig       `yaml:"service"`
	Orchestration OrchestrationConfig `yaml:"orchestration"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// ServiceConfig defines the basic parameters for the service to be managed.
type ServiceConfig struct {
	Name       string   `yaml:"name"`
	Command    []string `yaml:"command"`     // Main run command
	BinaryPath string   `yaml:"binary_path"` // Path for checks
	Env        []string `yaml:"env"`
}

// OrchestrationConfig defines the strategy and lifecycle hooks for process orchestration.
type OrchestrationConfig struct {
	Strategy     string             `yaml:"strategy"`
	PreFlight    []Hook             `yaml:"pre_flight"`   // Phase 1
	Startup      StartupConfig      `yaml:"startup"`      // Phase 2
	Canary       CanaryConfig       `yaml:"canary"`       // Phase 3
	Drain        DrainConfig        `yaml:"drain"`        // Phase 5
	PostProcess  PostProcessConfig  `yaml:"post_process"` // Phase 6
	StateHandoff StateHandoffConfig `yaml:"state_handoff"`
}

// Hook represents a custom command to be executed during specific orchestration phases.
type Hook struct {
	Name    string   `yaml:"name"`
	Command []string `yaml:"command"`
	Timeout string   `yaml:"timeout"`
}

// StartupConfig defines parameters for the process startup phase.
type StartupConfig struct {
	WarmupDelay string `yaml:"warmup_delay"`
}

// CanaryConfig defines parameters for the canary observation (soaking) phase.
type CanaryConfig struct {
	Enabled  bool   `yaml:"enabled"`
	SoakTime string `yaml:"soak_time"`
}

// DrainConfig defines parameters for the old process shutdown phase.
type DrainConfig struct {
	Timeout string `yaml:"timeout"`
}

// PostProcessConfig defines hooks to be executed after a success or failure of orchestration.
type PostProcessConfig struct {
	OnSuccess []Hook `yaml:"on_success"`
	OnFailure []Hook `yaml:"on_failure"`
}

// StateHandoffConfig defines parameters for the State Relay Protocol (SRP) memory context transfer.
type StateHandoffConfig struct {
	Enabled    bool   `yaml:"enabled"`
	SocketPath string `yaml:"socket_path"`
	Timeout    string `yaml:"timeout"`
}

// ObservabilityConfig defines parameters for metrics and logging.
type ObservabilityConfig struct {
	MetricsPort string `yaml:"metrics_port"`
	LogLevel    string `yaml:"log_level"`
}

// Personal.AI order the ending
