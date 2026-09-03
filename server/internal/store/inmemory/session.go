package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type inMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*models.Session
}

var _ store.SessionRepository = (*inMemorySessionRepository)(nil)

func newSessionRepository() *inMemorySessionRepository {
	return &inMemorySessionRepository{
		sessions: make(map[uuid.UUID]*models.Session),
	}
}

func NewSessionRepository() store.SessionRepository {
	return newSessionRepository()
}

func (r *inMemorySessionRepository) clone() *inMemorySessionRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemorySessionRepository{
		sessions: make(map[uuid.UUID]*models.Session, len(r.sessions)),
	}
	for id, s := range r.sessions {
		sc := *s
		c.sessions[id] = &sc
	}
	return c
}

func (r *inMemorySessionRepository) Create(session *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	} else if _, exists := r.sessions[session.ID]; exists {
		return fmt.Errorf("session with ID %s already exists", session.ID)
	}

	copy := *session
	r.sessions[session.ID] = &copy

	return nil
}

func (r *inMemorySessionRepository) FindByID(id uuid.UUID) (models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return models.Session{}, store.ErrNotFound
	}

	return *session, nil
}

func (r *inMemorySessionRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[id]; !exists {
		return store.ErrNotFound
	}

	delete(r.sessions, id)

	return nil
}

func (r *inMemorySessionRepository) DeleteByUsername(username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, session := range r.sessions {
		if session.Username == username {
			delete(r.sessions, id)
		}
	}

	return nil
}
