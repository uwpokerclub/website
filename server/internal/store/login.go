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

	// List retrieves all logins with their linked member information (matched by
	// quest ID), ordered by username ascending, optionally filtered by search across username,
	// role, and linked member name, along with the total matching count before pagination.
	List(pagination *models.Pagination, search string) ([]models.LoginWithMember, int64, error)

	// FindByUsernameWithMember retrieves a single login with its linked member information.
	// Returns store.ErrNotFound if no login exists for the given username.
	FindByUsernameWithMember(username string) (models.LoginWithMember, error)
}
