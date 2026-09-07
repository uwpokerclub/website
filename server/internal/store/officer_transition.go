package store

import "api/internal/models"

type OfficerTransitionRepository interface {
	Create(*models.OfficerTransition) error
	Current() (models.OfficerTransition, error)
}
