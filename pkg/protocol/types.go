package protocol

// Config represents the root configuration compliant with UPHR-O
type Config struct {
	Version       string              `yaml:"version"`
	Service       ServiceConfig       `yaml:"service"`
	Orchestration OrchestrationConfig `yaml:"orchestration"`
	Observability ObservabilityConfig `yaml:"observability"`
}

type ServiceConfig struct {
	Name       string   `yaml:"name"`
	Command    []string `yaml:"command"`     // Main run command
	BinaryPath string   `yaml:"binary_path"` // Path for checks
	Env        []string `yaml:"env"`
}

type OrchestrationConfig struct {
	Strategy     string             `yaml:"strategy"`
	PreFlight    []Hook             `yaml:"pre_flight"`   // Phase 1
	Startup      StartupConfig      `yaml:"startup"`      // Phase 2
	Canary       CanaryConfig       `yaml:"canary"`       // Phase 3
	Drain        DrainConfig        `yaml:"drain"`        // Phase 5
	PostProcess  PostProcessConfig  `yaml:"post_process"` // Phase 6
	StateHandoff StateHandoffConfig `yaml:"state_handoff"`
}

type Hook struct {
	Name    string   `yaml:"name"`
	Command []string `yaml:"command"`
	Timeout string   `yaml:"timeout"`
}

type StartupConfig struct {
	WarmupDelay string `yaml:"warmup_delay"`
}

type CanaryConfig struct {
	Enabled  bool   `yaml:"enabled"`
	SoakTime string `yaml:"soak_time"`
}

type DrainConfig struct {
	Timeout string `yaml:"timeout"`
}

type PostProcessConfig struct {
	OnSuccess []Hook `yaml:"on_success"`
	OnFailure []Hook `yaml:"on_failure"`
}

type StateHandoffConfig struct {
	Enabled    bool   `yaml:"enabled"`
	SocketPath string `yaml:"socket_path"`
	Timeout    string `yaml:"timeout"`
}

type ObservabilityConfig struct {
	MetricsPort string `yaml:"metrics_port"`
	LogLevel    string `yaml:"log_level"`
}

// Personal.AI order the ending
