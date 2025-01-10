package state

import "sync"

// ServiceDest is a struct that represents a service destination.
type ServiceDestState struct {
	// state is a map of service name to destination.
	// frp-service-1 -> 127.0.0.1:12345
	state          map[string]int
	mu             sync.RWMutex
	updateCallback func(map[string]int)
}

func NewServiceDestState(updateCallback func(map[string]int)) *ServiceDestState {
	if updateCallback == nil {
		updateCallback = func(state map[string]int) {}
	}
	return &ServiceDestState{
		state:          make(map[string]int),
		updateCallback: updateCallback,
	}
}
func (s *ServiceDestState) Set(service string, dest int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	go s.updateCallback(s.state)
}

func (s *ServiceDestState) SetMap(state map[string]int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.state = state
	go s.updateCallback(s.state)
}

func (s *ServiceDestState) Delete(service string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.state, service)
	go s.updateCallback(s.state)
}

func (s *ServiceDestState) Dump(template string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// TODO: Implement this function
	return template
}
