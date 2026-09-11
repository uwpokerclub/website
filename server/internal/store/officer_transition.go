package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

type OfficerTransitionRepository interface {
	Create(*models.OfficerTransition) error
	Current() (models.OfficerTransition, error)
	FindByIDForUpdate(uuid.UUID) (models.OfficerTransition, error)
	// ClaimCompletion atomically changes a pending transition to completed.
	// It returns ErrNotFound unless this caller won the guarded transition.
	ClaimCompletion(uuid.UUID) (models.OfficerTransition, error)
}
