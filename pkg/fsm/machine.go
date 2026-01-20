package fsm

import (
	"fmt"
	"sync"
)

// State represents a state in the finite state machine.
type State string

// Event represents a trigger that causes a state transition.
type Event string

// Handler is a function executed when a transition occurs.
// It receives the event that triggered the transition and any additional arguments.
type Handler func(event Event, args ...interface{}) error

// StateMachine is a thread-safe implementation of a finite state machine.
// It manages states, transitions, and callbacks associated with those transitions.
type StateMachine struct {
	mu          sync.RWMutex
	current     State
	transitions map[State]map[Event]State
	callbacks   map[State]map[Event]Handler
}

// New creates a new StateMachine with the specified initial state.
func New(initial State) *StateMachine {
	return &StateMachine{
		current:     initial,
		transitions: make(map[State]map[Event]State),
		callbacks:   make(map[State]map[Event]Handler),
	}
}

// Current returns the current state of the machine.
func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// AddTransition registers a transition from one state to another triggered by an event.
// It also associates a callback handler with the transition.
func (sm *StateMachine) AddTransition(from, to State, event Event, callback Handler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.transitions[from]; !ok {
		sm.transitions[from] = make(map[Event]State)
		sm.callbacks[from] = make(map[Event]Handler)
	}
	sm.transitions[from][event] = to
	sm.callbacks[from][event] = callback
}

// Fire triggers a state transition. It is thread-safe.
func (sm *StateMachine) Fire(event Event, args ...interface{}) error {
	sm.mu.Lock()

	stateTransitions, ok := sm.transitions[sm.current]
	if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("no transitions defined for state %s", sm.current)
	}

	next, ok := stateTransitions[event]
	if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("invalid transition from %s via %s", sm.current, event)
	}

	// Capture handler before changing state
	var handler Handler
	if handlers, exists := sm.callbacks[sm.current]; exists {
		handler = handlers[event]
	}

	sm.current = next
	sm.mu.Unlock()

	// Execute callback if exists
	if handler != nil {
		if err := handler(event, args...); err != nil {
			return err
		}
	}

	return nil
}

// Personal.AI order the ending
