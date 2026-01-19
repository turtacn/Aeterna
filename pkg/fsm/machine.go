package fsm

import (
	"fmt"
	"sync"
)

type State string
type Event string

// Handler is executed when a transition occurs
type Handler func(event Event, args ...interface{}) error

type StateMachine struct {
	mu          sync.RWMutex
	current     State
	transitions map[State]map[Event]State
	callbacks   map[State]map[Event]Handler
}

func New(initial State) *StateMachine {
	return &StateMachine{
		current:     initial,
		transitions: make(map[State]map[Event]State),
		callbacks:   make(map[State]map[Event]Handler),
	}
}

func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

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
	defer sm.mu.Unlock()

	next, ok := sm.transitions[sm.current][event]
	if !ok {
		return fmt.Errorf("invalid transition from %s via %s", sm.current, event)
	}

	// Execute callback if exists
	if handler, exists := sm.callbacks[sm.current][event]; exists && handler != nil {
		if err := handler(event, args...); err != nil {
			return err
		}
	}

	sm.current = next
	return nil
}

// Personal.AI order the ending
