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
	var total int64
	if err := r.db.Model(&models.Participant{}).
		Where("participants.event_id = ?", filter.EventID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := models.Participant{}.Preload(r.db).
		Where("participants.event_id = ?", filter.EventID).
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
