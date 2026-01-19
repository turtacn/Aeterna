package fsm

import (
	"fmt"
	"testing"
	"time"
)

func TestStateMachine_Deadlock(t *testing.T) {
	sm := New(State("initial"))

	sm.AddTransition(State("initial"), State("intermediate"), Event("first"), func(event Event, args ...interface{}) error {
		return sm.Fire(Event("second"))
	})

	sm.AddTransition(State("intermediate"), State("final"), Event("second"), nil)

	done := make(chan bool)
	go func() {
		err := sm.Fire(Event("first"))
		if err != nil {
			t.Errorf("Fire failed: %v", err)
		}
		done <- true
	}()

	select {
	case <-done:
		if sm.Current() != State("final") {
			t.Errorf("Expected state final, got %s", sm.Current())
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Deadlock detected: Fire did not return within 1 second")
	}
}

func TestStateMachine_Basic(t *testing.T) {
	sm := New(State("off"))
	sm.AddTransition(State("off"), State("on"), Event("push"), nil)

	if sm.Current() != State("off") {
		t.Errorf("Expected off, got %s", sm.Current())
	}

	err := sm.Fire(Event("push"))
	if err != nil {
		t.Fatal(err)
	}

	if sm.Current() != State("on") {
		t.Errorf("Expected on, got %s", sm.Current())
	}
}

func TestStateMachine_InvalidTransition(t *testing.T) {
	sm := New(State("start"))
	err := sm.Fire(Event("unknown"))
	if err == nil {
		t.Fatal("Expected error for unknown event")
	}
}

func TestStateMachine_HandlerError(t *testing.T) {
	sm := New(State("A"))
	sm.AddTransition(State("A"), State("B"), Event("go"), func(event Event, args ...interface{}) error {
		return fmt.Errorf("handler failed")
	})

	err := sm.Fire(Event("go"))
	if err == nil || err.Error() != "handler failed" {
		t.Fatalf("Expected handler failed error, got %v", err)
	}

	if sm.Current() != State("B") {
		t.Errorf("Expected state B even if handler failed, got %s", sm.Current())
	}
}

func TestStateMachine_StateConsistencyInHandler(t *testing.T) {
	sm := New(State("A"))
	var stateInHandler State
	sm.AddTransition(State("A"), State("B"), Event("go"), func(event Event, args ...interface{}) error {
		stateInHandler = sm.Current()
		return nil
	})

	sm.Fire(Event("go"))
	if stateInHandler != State("B") {
		t.Errorf("Expected handler to see state B, saw %s", stateInHandler)
	}
}
