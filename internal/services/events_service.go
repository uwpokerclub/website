package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type eventService struct {
	db *gorm.DB
}

func NewEventService(db *gorm.DB) *eventService {
	return &eventService{
		db: db,
	}
}

func (es *eventService) CreateEvent(req *models.CreateEventRequest) (*models.Event, error) {
	semesterId, err := uuid.Parse(req.SemesterID)
	if err != nil {
		return nil, e.InvalidRequest("Invalid semester ID specified in request")
	}

	event := models.Event{
		Name:        req.Name,
		Format:      req.Format,
		Notes:       req.Notes,
		SemesterID:  semesterId,
		StartDate:   req.StartDate,
		State:       models.EventStateStarted,
		StructureID: req.StructureID,
		Rebuys:      0,
	}

	res := es.db.Create(&event)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &event, nil
}

func (es *eventService) GetEvent(eventId uint64) (*models.Event, error) {
	event := models.Event{ID: eventId}

	res := es.db.First(&event)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &event, nil
}

func (es *eventService) ListEvents(semesterId string) ([]models.Event, error) {
	var events []models.Event

	query := es.db.
		Table("events").
		Joins(
			"LEFT JOIN (?) as entries ON events.id = entries.event_id",
			es.db.Table("participants").Select("event_id, COUNT(*)").Group("event_id"),
		)

	if semesterId != "" {
		semesterUUID, err := uuid.Parse(semesterId)
		if err != nil {
			return nil, e.InvalidRequest("Invalid semester ID specified in request")
		}

		query = query.Where("events.semester_id = ?", semesterUUID)
	}

	res := query.Order("events.start_date DESC").Scan(&events)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return events, nil
}

func (es *eventService) EndEvent(eventId uint64) error {
	// Retrieve the event first
	event := models.Event{ID: eventId}
	res := es.db.First(&event)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	// Events cannot be ended twice, so reject any request attempting to do this
	if event.State == models.EventStateEnded {
		return e.Forbidden("This event has already ended, it cannot be ended again.")
	}

	// Start transaction for the event update and ranking update process
	tx := es.db.Begin()
	if err := tx.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	// Update all entries who are not signed out to sign them out in last place
	res = tx.Model(&models.Participant{}).Where("event_id = ? AND signed_out_at IS NULL", event.ID).Update("signed_out_at", event.StartDate)
	if err := res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	// Update events state
	event.State = models.EventStateEnded
	tx.Save(&event)
	if err := tx.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	// Retrieve list of entries for the event
	entries := []models.Participant{}
	res = tx.Where("event_id = ?", eventId).Order("signed_out_at DESC").Find(&entries)
	if err := res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	// Calculate points for each entry in the event and update placements
	rankingService := NewRankingService(tx)
	eventSize := len(entries)
	for i, entry := range entries {
		points := CalculatePoints(eventSize, i+1)

		err := rankingService.UpdateRanking(entry.MembershipID, points)
		if err != nil {
			tx.Rollback()
			return err
		}

		entry.Placement = uint32(i + 1)

		tx.Save(&entry)
		if err := tx.Error; err != nil {
			tx.Rollback()
			return e.InternalServerError(err.Error())
		}
	}

	// Save all changes to the database
	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (es *eventService) NewRebuy(eventId uint64) error {
	event := models.Event{ID: eventId}
	res := es.db.First(&event)

	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return e.NotFound(err.Error())
	} else if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateEnded {
		return e.Forbidden("This event has already ended, it cannot be ended again.")
	}

	tx := es.db.Begin()
	if err := tx.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	event.Rebuys += 1

	semesterService := NewSemesterService(tx)

	semester, err := semesterService.GetSemester(event.SemesterID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = semesterService.UpdateBudget(event.SemesterID, float64(semester.RebuyFee))
	if err != nil {
		tx.Rollback()
		return err
	}

	res = tx.Save(&event)
	if err := res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	return nil
}
