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

// FindByEventIDForUpdate has no real locking mechanism in the in-memory
// store, and unlike a genuine SELECT ... FOR UPDATE it does not block a
// concurrent transaction: BeginTx snapshots the store the moment it is
// called, without waiting on any other open transaction, and Commit later
// overwrites the parent's repos with that snapshot's contents. Two
// transactions that both call BeginTx before either commits therefore each
// mutate their own stale clone, and whichever commits second silently wins,
// discarding the first transaction's write - a lost update a real row lock
// would have prevented by blocking the second until the first committed.
// This is a pre-existing characteristic of InMemoryStore.BeginTx/Commit
// shared by every repository, not something specific to event clocks or
// fixable here; production correctness for the transactional action path
// relies on the postgres implementation's real lock (see
// TestEventClockRepository_FindByEventIDForUpdate_BlocksConcurrentTx in
// store/postgres).
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
