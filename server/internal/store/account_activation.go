package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

type AccountActivationRepository interface {
	Create(activation *models.AccountActivation) error
	DeleteUnusedByUsername(username string) error
	DeleteByTransition(id uuid.UUID) error
	FindValid(tokenHash []byte) (models.AccountActivation, error)
	Consume(tokenHash []byte) (models.AccountActivation, error)
}
