package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

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

	query := applyFilter(models.Participant{}.Preload(r.db)).
		Order("participants.signed_out_at DESC")
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
