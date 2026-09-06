package store

import "api/internal/models"

type AccountActivationRepository interface {
	Create(activation *models.AccountActivation) error
	DeleteUnusedByUsername(username string) error
	FindValid(tokenHash []byte) (models.AccountActivation, error)
	Consume(tokenHash []byte) (models.AccountActivation, error)
}
