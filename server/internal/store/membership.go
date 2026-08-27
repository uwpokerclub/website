package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// MembershipRepository is the interface for accessing the memberships in the data store. It provides methods for creating, reading, updating, and deleting memberships.
type MembershipRepository interface {
	// Create creates a new membership in the data store.
	Create(membership *models.Membership) error

	// FindByID retrieves a membership from the data store by its ID, preloaded with its User
	// and Semester.
	FindByID(id uuid.UUID) (models.Membership, error)

	// FindByIDAndSemesterID retrieves a membership from the data store by its ID, scoped to a
	// specific semester, preloaded with its User and Semester.
	FindByIDAndSemesterID(id uuid.UUID, semesterID uuid.UUID) (models.Membership, error)

	// List retrieves memberships matching filter (SemesterID, UserID, Paid, Discounted, and the
	// joined-user filters Search/Name/Email/Faculty/StudentID), each with its computed
	// event-attendance count within filter.SemesterID, ordered by first/last name ascending,
	// along with the total matching count before pagination is applied.
	List(filter *models.ListMembershipsFilter) ([]models.MembershipWithAttendance, int64, error)

	// Update updates an existing membership's Paid and Discounted fields in the data store.
	Update(membership *models.Membership) error

	// Delete deletes a membership from the data store by its ID, scoped to a specific semester. Returns
	// ErrNotFound if no matching record exists.
	Delete(id uuid.UUID, semesterID uuid.UUID) error

	// SetFreeTrialAvailable atomically sets a membership's free trial availability flag via a
	// single UPDATE, not a read-modify-write, so it cannot clobber a concurrent change to
	// Paid/Discounted (mirrors SemesterRepository.IncrementBudget). Returns store.ErrNotFound if
	// no membership exists for the given ID.
	SetFreeTrialAvailable(id uuid.UUID, available bool) error
}
