package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// SemesterRepository is the interface for accessing the semesters in the data store. It provides methods for creating, reading, updating, and deleting semesters.
type SemesterRepository interface {
	Create(semester *models.Semester) error
	FindByID(id uuid.UUID) (models.Semester, error)
	List(pagination *models.Pagination) ([]models.Semester, int64, error)
	Update(semester *models.Semester) error

	// IncrementBudget atomically adds amount to a semester's current budget (amount may be
	// negative to subtract). This is a single atomic SQL update, not a read-modify-write, to
	// avoid lost updates under concurrent budget changes. Returns store.ErrNotFound if no
	// semester exists for the given ID.
	IncrementBudget(id uuid.UUID, amount float32) error
}
