package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresSessionRepository struct {
	db *gorm.DB
}

var _ store.SessionRepository = (*postgresSessionRepository)(nil)

func NewSessionRepository(db *gorm.DB) store.SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *postgresSessionRepository) FindByID(id uuid.UUID) (models.Session, error) {
	var session models.Session

	err := r.db.Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Session{}, store.ErrNotFound
		}
		return models.Session{}, err
	}

	return session, nil
}

func (r *postgresSessionRepository) Delete(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.Session{})
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *postgresSessionRepository) DeleteByUsername(username string) error {
	return r.db.Where("username = ?", username).Delete(&models.Session{}).Error
}
