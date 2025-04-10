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
	Pending   bool
	Hash      string
	MaxLength int
}

type PendingTask struct {
	ID        string
	Hash      string
	MaxLength int
	Status    string
}

type RequestStore interface {
	Get(id string) (RequestState, bool)
	Update(id string, state RequestState)
	Set(id string, state RequestState)
	Count() int
	MarkPending(id string, isPending bool)
	GetPending() []PendingTask
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

func (r *requestStoreImpl) MarkPending(id string, isPending bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, ok := r.store[id]
	if !ok {
		return
	}
	state.Pending = isPending
	r.store[id] = state
}

func (r *requestStoreImpl) GetPending() []PendingTask {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []PendingTask
	for id, s := range r.store {
		if s.Pending {
			result = append(result, PendingTask{
				ID:        id,
				Hash:      s.Hash,
				MaxLength: s.MaxLength,
				Status:    s.Status,
			})
		}
	}
	return result
}
