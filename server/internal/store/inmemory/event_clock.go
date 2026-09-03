package inmemory

import (
	"sync"

	"api/internal/models"
	"api/internal/store"
)

type inMemoryEventClockRepository struct {
	mu     sync.RWMutex
	clocks map[int32]*models.EventClock
}

var _ store.EventClockRepository = (*inMemoryEventClockRepository)(nil)

func newEventClockRepository() *inMemoryEventClockRepository {
	return &inMemoryEventClockRepository{
		clocks: make(map[int32]*models.EventClock),
	}
}

func NewEventClockRepository() store.EventClockRepository {
	return newEventClockRepository()
}

func (r *inMemoryEventClockRepository) clone() *inMemoryEventClockRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryEventClockRepository{
		clocks: make(map[int32]*models.EventClock, len(r.clocks)),
	}
	for id, clock := range r.clocks {
		cc := *clock
		c.clocks[id] = &cc
	}
	return c
}

func (r *inMemoryEventClockRepository) FindByEventID(eventID int32) (models.EventClock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clock, exists := r.clocks[eventID]
	if !exists {
		return models.EventClock{}, store.ErrNotFound
	}

	return *clock, nil
}

// FindByEventIDForUpdate has no separate locking mechanism in the in-memory
// store: BeginTx already snapshots the whole store under a read lock, and
// Commit swaps the parent's repos under a write lock, so a transaction's view
// is already isolated from concurrent transactions the same way a real
// SELECT ... FOR UPDATE isolates a row.
func (r *inMemoryEventClockRepository) FindByEventIDForUpdate(eventID int32) (models.EventClock, error) {
	return r.FindByEventID(eventID)
}

func (r *inMemoryEventClockRepository) Create(clock *models.EventClock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clocks[clock.EventID]; exists {
		return store.ErrAlreadyExists
	}

	copy := *clock
	r.clocks[clock.EventID] = &copy

	return nil
}

func (r *inMemoryEventClockRepository) Update(clock *models.EventClock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clocks[clock.EventID]; !exists {
		return store.ErrNotFound
	}

	copy := *clock
	r.clocks[clock.EventID] = &copy

	return nil
}

func (r *inMemoryEventClockRepository) DeleteByEventID(eventID int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clocks[eventID]; !exists {
		return store.ErrNotFound
	}

	delete(r.clocks, eventID)

	return nil
}
