package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// SessionRepository is the interface for accessing the sessions in the data store. It provides
// methods for creating, reading, and deleting sessions.
type SessionRepository interface {
	// Create creates a new session in the data store.
	Create(session *models.Session) error

	// FindByID retrieves a session from the data store by its ID.
	// Returns store.ErrNotFound if no session exists for the given ID.
	FindByID(id uuid.UUID) (models.Session, error)

	// Delete deletes a session from the data store by its ID.
	// Returns store.ErrNotFound if no session exists for the given ID.
	Delete(id uuid.UUID) error
}
