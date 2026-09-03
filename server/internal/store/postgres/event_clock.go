package postgres

import (
	"errors"

	"api/internal/models"
	"api/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postgresEventClockRepository struct {
	db *gorm.DB
}

var _ store.EventClockRepository = (*postgresEventClockRepository)(nil)

func NewEventClockRepository(db *gorm.DB) store.EventClockRepository {
	return &postgresEventClockRepository{db: db}
}

func (r *postgresEventClockRepository) FindByEventID(eventID int32) (models.EventClock, error) {
	return r.find(r.db, eventID)
}

func (r *postgresEventClockRepository) FindByEventIDForUpdate(eventID int32) (models.EventClock, error) {
	return r.find(r.db.Clauses(clause.Locking{Strength: "UPDATE"}), eventID)
}

func (r *postgresEventClockRepository) find(tx *gorm.DB, eventID int32) (models.EventClock, error) {
	var clock models.EventClock

	res := tx.First(&clock, "event_id = ?", eventID)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.EventClock{}, store.ErrNotFound
		}
		return models.EventClock{}, err
	}

	return clock, nil
}

func (r *postgresEventClockRepository) Create(clock *models.EventClock) error {
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(clock)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return store.ErrAlreadyExists
	}

	return nil
}

func (r *postgresEventClockRepository) Update(clock *models.EventClock) error {
	result := r.db.Model(&models.EventClock{}).
		Where("event_id = ?", clock.EventID).
		Select("level_index", "level_ends_at", "paused_at", "version", "updated_at").
		Updates(clock)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *postgresEventClockRepository) DeleteByEventID(eventID int32) error {
	result := r.db.Delete(&models.EventClock{}, "event_id = ?", eventID)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}
