package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"encoding/hex"
	"sync"
	"time"
)

type inMemoryAccountActivationRepository struct {
	mu          sync.RWMutex
	activations map[string]*models.AccountActivation
}

var _ store.AccountActivationRepository = (*inMemoryAccountActivationRepository)(nil)

func newAccountActivationRepository() *inMemoryAccountActivationRepository {
	return &inMemoryAccountActivationRepository{activations: map[string]*models.AccountActivation{}}
}
func (r *inMemoryAccountActivationRepository) clone() *inMemoryAccountActivationRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c := newAccountActivationRepository()
	for key, activation := range r.activations {
		copy := *activation
		copy.TokenHash = append([]byte(nil), activation.TokenHash...)
		c.activations[key] = &copy
	}
	return c
}
func (r *inMemoryAccountActivationRepository) Create(a *models.AccountActivation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := hex.EncodeToString(a.TokenHash)
	copy := *a
	copy.TokenHash = append([]byte(nil), a.TokenHash...)
	r.activations[key] = &copy
	return nil
}
func (r *inMemoryAccountActivationRepository) DeleteUnusedByUsername(username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, activation := range r.activations {
		if activation.Username == username && activation.UsedAt == nil {
			delete(r.activations, key)
		}
	}
	return nil
}
func (r *inMemoryAccountActivationRepository) FindValid(hash []byte) (models.AccountActivation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.activations[hex.EncodeToString(hash)]
	if !ok || a.UsedAt != nil || !a.ExpiresAt.After(time.Now().UTC()) {
		return models.AccountActivation{}, store.ErrNotFound
	}
	return *a, nil
}
func (r *inMemoryAccountActivationRepository) Consume(hash []byte) (models.AccountActivation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.activations[hex.EncodeToString(hash)]
	if !ok || a.UsedAt != nil || !a.ExpiresAt.After(time.Now().UTC()) {
		return models.AccountActivation{}, store.ErrNotFound
	}
	now := time.Now().UTC()
	a.UsedAt = &now
	return *a, nil
}
