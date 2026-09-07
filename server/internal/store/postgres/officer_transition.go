package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type postgresOfficerTransitionRepository struct{ db *gorm.DB }

var _ store.OfficerTransitionRepository = (*postgresOfficerTransitionRepository)(nil)

func NewOfficerTransitionRepository(db *gorm.DB) store.OfficerTransitionRepository {
	return &postgresOfficerTransitionRepository{db}
}
func (r *postgresOfficerTransitionRepository) Create(t *models.OfficerTransition) error {
	return mapOfficerTransitionCreateError(r.db.Create(t).Error)
}

func mapOfficerTransitionCreateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "one_pending_transition" {
		return store.ErrConflict
	}
	return err
}
func (r *postgresOfficerTransitionRepository) Current() (models.OfficerTransition, error) {
	var t models.OfficerTransition
	err := r.db.Where("status = ?", models.OfficerTransitionPending).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.OfficerTransition{}, store.ErrNotFound
	}
	return t, err
}
