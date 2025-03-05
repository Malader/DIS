package store

import (
	"sync"
	"time"
)

type RequestState struct {
	Status    string
	Data      []string
	Timer     *time.Timer
	StartTime time.Time
	Timeout   time.Duration
}

type RequestStore interface {
	Get(id string) (RequestState, bool)
	Update(id string, state RequestState)
	Set(id string, state RequestState)
	Count() int
}

type requestStoreImpl struct {
	mu    sync.RWMutex
	store map[string]RequestState
}

func NewRequestStore() RequestStore {
	return &requestStoreImpl{
		store: make(map[string]RequestState),
	}
}

func (r *requestStoreImpl) Set(id string, state RequestState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[id] = state
}

func (r *requestStoreImpl) Get(id string) (RequestState, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.store[id]
	return s, ok
}

func (r *requestStoreImpl) Update(id string, state RequestState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[id] = state
}

func (r *requestStoreImpl) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.store)
}
