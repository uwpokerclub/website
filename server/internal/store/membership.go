package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// MembershipRepository is the interface for accessing the memberships in the data store. It provides methods for creating, reading, updating, and deleting memberships.
type MembershipRepository interface {
	// Create creates a new membership in the data store.
	Create(membership *models.Membership) error

	// FindByID retrieves a membership from the data store by its ID.
	FindByID(id uuid.UUID) (models.Membership, error)

	// FindByIDAndSemesterID retrieves a membership from the data store by its ID, scoped to a specific semester.
	FindByIDAndSemesterID(id uuid.UUID, semesterID uuid.UUID) (models.Membership, error)

	// List retrieves memberships from the data store matching the filter's SemesterID, UserID, Paid, and
	// Discounted fields, ordered by UserID ascending. Filtering on joined member fields (name, email,
	// faculty, student ID, search) requires a join against the members table and is not performed at
	// this layer; callers needing that filtering continue to use the service layer for now.
	List(filter *models.ListMembershipsFilter) ([]models.Membership, int64, error)

	// Update updates an existing membership's Paid and Discounted fields in the data store.
	Update(membership *models.Membership) error

	// Delete deletes a membership from the data store by its ID, scoped to a specific semester. Returns
	// ErrNotFound if no matching record exists.
	Delete(id uuid.UUID, semesterID uuid.UUID) error
}
