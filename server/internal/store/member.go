package store

import "api/internal/models"

// MemberRepository is the interface for accessing the members in the data store. It provides methods for creating, reading, updating, and deleting logins.
type MemberRepository interface {
	// Create creates a new member in the data store.
	Create(member *models.User) error

	// FindByID finds a member by their ID. It returns the member or error encountered.
	FindByID(id uint64) (*models.User, error)

	// List returns a list of all members in the data store matching the given filter. It returns the list of members or error encountered.
	List(filter *models.ListUsersFilter, pagination *models.Pagination) ([]*models.User, int64, error)

	// Update updates an existing member in the data store. It returns error encountered.
	Update(member *models.User) error

	// Delete deletes a member from the data store by their ID. Returns ErrNotFound if no record exists.
	Delete(id uint64) error
}
