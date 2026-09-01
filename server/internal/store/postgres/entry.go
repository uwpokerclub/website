package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postgresEntryRepository struct {
	db *gorm.DB
}

var _ store.EntryRepository = (*postgresEntryRepository)(nil)

func NewEntryRepository(db *gorm.DB) store.EntryRepository {
	return &postgresEntryRepository{db: db}
}

func (r *postgresEntryRepository) Create(participant *models.Participant) error {
	return r.db.Create(participant).Error
}

func (r *postgresEntryRepository) FindByID(id int32) (models.Participant, error) {
	var participant models.Participant

	if err := r.db.First(&participant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Participant{}, store.ErrNotFound
		}
		return models.Participant{}, err
	}

	return participant, nil
}

func (r *postgresEntryRepository) FindByMembershipAndEventID(membershipID uuid.UUID, eventID int32) (models.Participant, error) {
	var participant models.Participant

	err := r.db.
		Where("membership_id = ? AND event_id = ?", membershipID, eventID).
		First(&participant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Participant{}, store.ErrNotFound
		}
		return models.Participant{}, err
	}

	return participant, nil
}

func (r *postgresEntryRepository) List(filter *models.ListParticipantsFilter) ([]models.Participant, int64, error) {
	applyFilter := func(q *gorm.DB) *gorm.DB {
		q = q.Where("participants.event_id = ?", filter.EventID)
		if filter.Search != "" {
			pattern := "%" + eventNameLikeReplacer.Replace(filter.Search) + "%"
			q = q.Joins("JOIN memberships ON memberships.id = participants.membership_id").
				Joins("JOIN users ON users.id = memberships.user_id").
				Where(
					"users.first_name ILIKE ? OR users.last_name ILIKE ? OR (users.first_name || ' ' || users.last_name) ILIKE ? OR CAST(users.id AS TEXT) ILIKE ?",
					pattern, pattern, pattern, pattern,
				)
		}
		return q
	}

	var total int64
	if err := applyFilter(r.db.Model(&models.Participant{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// The id tiebreaker makes the order total: SignOutAllUnsigned gives every still-seated entry
	// one shared timestamp, and without it LIMIT/OFFSET can repeat or skip rows across pages.
	// inmemory/entry.go sorts to match.
	query := applyFilter(models.Participant{}.Preload(r.db)).
		Order("participants.signed_out_at DESC, participants.id DESC")
	query = filter.Pagination.Apply(query)

	var participants []models.Participant
	if err := query.Find(&participants).Error; err != nil {
		return nil, 0, err
	}

	return participants, total, nil
}

func (r *postgresEntryRepository) Update(participant *models.Participant, values map[string]any) error {
	return r.db.Omit(clause.Associations).Model(participant).Updates(values).Error
}

func (r *postgresEntryRepository) Delete(membershipID uuid.UUID, eventID int32) error {
	result := r.db.
		Where("event_id = ?", eventID).
		Delete(&models.Participant{}, "membership_id = ?", membershipID)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *postgresEntryRepository) SignOutAllUnsigned(eventID int32, signedOutAt time.Time) error {
	return r.db.Model(&models.Participant{}).
		Where("event_id = ? AND signed_out_at IS NULL", eventID).
		Update("signed_out_at", signedOutAt).Error
}

func (r *postgresEntryRepository) BatchUpdatePoints(points map[int32]int32) error {
	if len(points) == 0 {
		return nil
	}

	ids := make([]int32, 0, len(points))
	for id := range points {
		ids = append(ids, id)
	}

	caseExprs := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids)*2)
	argIdx := 1
	for _, id := range ids {
		caseExprs = append(caseExprs, fmt.Sprintf("WHEN id = $%d THEN $%d::integer", argIdx, argIdx+1))
		args = append(args, id, points[id])
		argIdx += 2
	}

	idPlaceholders := make([]string, 0, len(ids))
	for _, id := range ids {
		idPlaceholders = append(idPlaceholders, fmt.Sprintf("$%d", argIdx))
		args = append(args, id)
		argIdx++
	}

	query := fmt.Sprintf(
		"UPDATE participants SET points = CASE %s END WHERE id IN (%s)",
		strings.Join(caseExprs, " "),
		strings.Join(idPlaceholders, ", "),
	)

	return r.db.Exec(query, args...).Error
}

func (r *postgresEntryRepository) CountByMembershipID(membershipID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Participant{}).Where("membership_id = ?", membershipID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
