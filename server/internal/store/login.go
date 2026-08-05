package store

import "api/internal/models"

// LoginRepository is the interface for accessing the logins in the data store. It provides
// methods for creating, reading, updating, and deleting logins.
type LoginRepository interface {
	// Create creates a new login in the data store.
	Create(login *models.Login) error

	// FindByUsername retrieves a login from the data store by its username.
	// Returns store.ErrNotFound if no login exists for the given username.
	FindByUsername(username string) (models.Login, error)

	// Update applies a partial update to a login using the given column/value map.
	// Returns store.ErrNotFound if no login exists for the given username.
	Update(username string, values map[string]any) error

	// Delete deletes a login from the data store by its username.
	// Returns store.ErrNotFound if no login exists for the given username.
	Delete(username string) error
}
