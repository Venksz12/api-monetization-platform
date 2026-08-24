package idempotency

import (
	"sync"
)

type Store struct {
	mu sync.Mutex
	m  map[string]any
}

func New() *Store {
	return &Store{m: make(map[string]any)}
}

func (s *Store) Get(key string) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[key]
	return v, ok
}

func (s *Store) Put(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}
