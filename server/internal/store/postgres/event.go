package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postgresEventRepository struct {
	db *gorm.DB
}

var _ store.EventRepository = (*postgresEventRepository)(nil)

func NewEventRepository(db *gorm.DB) store.EventRepository {
	return &postgresEventRepository{db: db}
}

// eventNameLikeReplacer escapes ILIKE special characters in user-supplied
// search input, mirroring services.sanitizeLikeInput.
var eventNameLikeReplacer = strings.NewReplacer(
	"\\", "\\\\",
	"%", "\\%",
	"_", "\\_",
)

func (r *postgresEventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *postgresEventRepository) FindByID(id int32) (models.Event, error) {
	event := models.Event{ID: id}

	res := event.Preload(r.db, models.EventPreloadOptions{
		Semester:  true,
		Structure: true,
		Entries:   true,
	}).First(&event)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Event{}, store.ErrNotFound
		}
		return models.Event{}, err
	}

	return event, nil
}

func (r *postgresEventRepository) FindBySemesterAndID(semesterID uuid.UUID, id int32) (models.Event, error) {
	var event models.Event

	query := models.Event{}.Preload(r.db, models.EventPreloadOptions{
		Semester:  true,
		Structure: true,
		Entries:   true,
	}).
		Where(`"Semester"."id" = ?`, semesterID).
		Where("events.id = ?", id)

	res := query.First(&event)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Event{}, store.ErrNotFound
		}
		return models.Event{}, err
	}

	return event, nil
}

func (r *postgresEventRepository) List(filter *models.ListEventsFilter) ([]models.Event, int64, error) {
	applyFilter := func(q *gorm.DB) *gorm.DB {
		if filter.SemesterID != nil {
			q = q.Where("semester_id = ?", *filter.SemesterID)
		}
		if filter.Search != "" {
			pattern := "%" + eventNameLikeReplacer.Replace(filter.Search) + "%"
			q = q.Where("name ILIKE ?", pattern)
		}
		return q
	}

	var total int64
	if err := applyFilter(r.db.Model(&models.Event{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := applyFilter(models.Event{}.Preload(r.db, models.EventPreloadOptions{Entries: true})).
		Order("start_date DESC")
	query = filter.Pagination.Apply(query)

	var events []models.Event
	if err := query.Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *postgresEventRepository) Update(event *models.Event, values map[string]any) error {
	return r.db.Omit(clause.Associations).Model(event).Updates(values).Error
}

func (r *postgresEventRepository) Delete(id int32) error {
	result := r.db.Delete(&models.Event{}, "id = ?", id)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}
