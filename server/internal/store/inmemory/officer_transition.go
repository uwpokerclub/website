package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"sync"
)

type inMemoryOfficerTransitionRepository struct {
	mu          sync.RWMutex
	transitions map[string]*models.OfficerTransition
}

var _ store.OfficerTransitionRepository = (*inMemoryOfficerTransitionRepository)(nil)

func newOfficerTransitionRepository() *inMemoryOfficerTransitionRepository {
	return &inMemoryOfficerTransitionRepository{transitions: map[string]*models.OfficerTransition{}}
}
func (r *inMemoryOfficerTransitionRepository) clone() *inMemoryOfficerTransitionRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c := newOfficerTransitionRepository()
	for k, v := range r.transitions {
		x := *v
		c.transitions[k] = &x
	}
	return c
}
func (r *inMemoryOfficerTransitionRepository) Create(t *models.OfficerTransition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.transitions {
		if x.Status == models.OfficerTransitionPending {
			return store.ErrConflict
		}
	}
	x := *t
	r.transitions[t.ID.String()] = &x
	return nil
}
func (r *inMemoryOfficerTransitionRepository) Current() (models.OfficerTransition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.transitions {
		if t.Status == models.OfficerTransitionPending {
			return *t, nil
		}
	}
	return models.OfficerTransition{}, store.ErrNotFound
}
