package orchestrator

import (
	"testing"
	"github.com/turtacn/Aeterna/pkg/protocol"
	"github.com/turtacn/Aeterna/pkg/consts"
	"github.com/turtacn/Aeterna/pkg/fsm"
)

func TestNewEngine(t *testing.T) {
	cfg := &protocol.Config{
		Service: protocol.ServiceConfig{
			Name: "test-service",
		},
		Orchestration: protocol.OrchestrationConfig{
			StateHandoff: protocol.StateHandoffConfig{
				SocketPath: "/tmp/aeterna-test.sock",
			},
		},
	}
	engine := NewEngine(cfg)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if engine.fsm.Current() != fsm.State(consts.StatePending) {
		t.Errorf("Expected state %v, got %v", consts.StatePending, engine.fsm.Current())
	}
}

func TestEngine_SetupFSM(t *testing.T) {
	cfg := &protocol.Config{}
	_ = NewEngine(cfg)
	// Transitions are set up in NewEngine
}

func TestEngine_InitialState(t *testing.T) {
	cfg := &protocol.Config{
		Orchestration: protocol.OrchestrationConfig{
			Canary: protocol.CanaryConfig{
				SoakTime: "100ms",
			},
		},
	}
	e := NewEngine(cfg)

	// Initial state
	if e.fsm.Current() != fsm.State(consts.StatePending) {
		t.Errorf("Expected PENDING, got %v", e.fsm.Current())
	}
}
