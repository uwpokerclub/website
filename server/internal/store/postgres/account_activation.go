package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type postgresAccountActivationRepository struct{ db *gorm.DB }

var _ store.AccountActivationRepository = (*postgresAccountActivationRepository)(nil)

func NewAccountActivationRepository(db *gorm.DB) store.AccountActivationRepository {
	return &postgresAccountActivationRepository{db: db}
}

func (r *postgresAccountActivationRepository) Create(activation *models.AccountActivation) error {
	return r.db.Create(activation).Error
}

func (r *postgresAccountActivationRepository) DeleteUnusedByUsername(username string) error {
	return r.db.Where("username = ? AND used_at IS NULL", username).Delete(&models.AccountActivation{}).Error
}

func (r *postgresAccountActivationRepository) DeleteByTransition(id uuid.UUID) error {
	return r.db.Where("transition_id = ?", id).Delete(&models.AccountActivation{}).Error
}

func (r *postgresAccountActivationRepository) FindValid(tokenHash []byte) (models.AccountActivation, error) {
	var activation models.AccountActivation
	err := r.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now().UTC()).First(&activation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AccountActivation{}, store.ErrNotFound
	}
	return activation, err
}

func (r *postgresAccountActivationRepository) Consume(tokenHash []byte) (models.AccountActivation, error) {
	var activation models.AccountActivation
	result := r.db.Raw(`UPDATE account_activations SET used_at = CURRENT_TIMESTAMP
		WHERE token_hash = ? AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP
		RETURNING token_hash, username, expires_at, used_at, transition_id`, tokenHash).Scan(&activation)
	if result.Error != nil {
		return models.AccountActivation{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.AccountActivation{}, store.ErrNotFound
	}
	return activation, nil
}
