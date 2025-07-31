package monzo

import (
	"sync"
)

type stateStore struct {
	mu    sync.RWMutex
	state string
}

func (s *stateStore) get() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state
}

func (s *stateStore) set(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = state
}

func (s *stateStore) delete() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = ""
}
