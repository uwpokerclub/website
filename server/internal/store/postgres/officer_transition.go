package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *postgresOfficerTransitionRepository) FindByIDForUpdate(id uuid.UUID) (models.OfficerTransition, error) {
	var transition models.OfficerTransition
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&transition, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.OfficerTransition{}, store.ErrNotFound
	}
	return transition, err
}

func (r *postgresOfficerTransitionRepository) ClaimCompletion(id uuid.UUID) (models.OfficerTransition, error) {
	var transition models.OfficerTransition
	result := r.db.Raw(`UPDATE officer_transitions SET status = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = ?
		RETURNING id, initiated_by, president_username, vice_president_username, secretary_username, treasurer_username, status, created_at, resolved_at`,
		"completed", id, models.OfficerTransitionPending).Scan(&transition)
	if result.Error != nil {
		return models.OfficerTransition{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.OfficerTransition{}, store.ErrNotFound
	}
	return transition, nil
}
