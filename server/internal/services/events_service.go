package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"errors"

	"github.com/google/uuid"
)

type eventService struct {
	store store.Store
}

func NewEventService(st store.Store) *eventService {
	return &eventService{
		store: st,
	}
}

func (svc *eventService) UpdateEvent(eventID int32, req *models.UpdateEventRequest) (*models.Event, error) {
	event, err := svc.store.Events().FindByID(eventID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, e.NotFound(err.Error())
		}
		return nil, e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateEnded {
		return nil, e.Forbidden("This event has ended. It cannot be updated.")
	}

	values := map[string]any{}
	if req.Name != nil {
		values["name"] = *req.Name
	}
	if req.Format != nil {
		values["format"] = *req.Format
	}
	if req.Notes != nil {
		values["notes"] = *req.Notes
	}
	if req.StartDate != nil {
		values["start_date"] = *req.StartDate
	}
	if req.PointsMultiplier != nil {
		values["points_multiplier"] = *req.PointsMultiplier
	}

	if err := svc.store.Events().Update(&event, values); err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &event, nil
}

func (svc *eventService) EndEvent(eventId int32) error {
	event, err := svc.store.Events().FindByID(eventId)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return e.NotFound(err.Error())
		}
		return e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateEnded {
		return e.Forbidden("This event has already ended, it cannot be ended again.")
	}

	tx, err := svc.store.BeginTx()
	if err != nil {
		return e.InternalServerError(err.Error())
	}
	defer tx.Rollback()

	if err := tx.Entries().SignOutAllUnsigned(event.ID, event.StartDate); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Events().Update(&event, map[string]any{"state": models.EventStateEnded}); err != nil {
		return e.InternalServerError(err.Error())
	}

	// Every entry now has a non-nil SignedOutAt (the step above set it for anyone still signed
	// in), so List's existing "signed_out_at DESC" order is exactly the placement order: the
	// most-recently-signed-out entry (the last one remaining, or the one force-ended) is placed
	// first.
	entries, _, err := tx.Entries().List(&models.ListParticipantsFilter{EventID: eventId})
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	eventSize := len(entries)
	rankingUpdates := make(map[uuid.UUID]int32, eventSize)
	placements := make(map[int32]uint16, eventSize)
	for i, entry := range entries {
		placement := i + 1
		placements[entry.ID] = uint16(placement)

		if entry.MembershipID != nil {
			points := CalculatePoints(eventSize, placement, event.PointsMultiplier)
			rankingUpdates[*entry.MembershipID] = int32(points)
		}
	}

	if err := tx.Entries().BatchUpdatePlacements(placements); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Rankings().BatchIncrementPoints(rankingUpdates); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *eventService) UndoEndEvent(eventId int32) error {
	event, err := svc.store.Events().FindByID(eventId)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return e.NotFound(err.Error())
		}
		return e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateStarted {
		return e.Forbidden("This event has not been ended")
	}

	tx, err := svc.store.BeginTx()
	if err != nil {
		return e.InternalServerError(err.Error())
	}
	defer tx.Rollback()

	if err := tx.Events().Update(&event, map[string]any{"state": models.EventStateStarted}); err != nil {
		return e.InternalServerError(err.Error())
	}

	entries, _, err := tx.Entries().List(&models.ListParticipantsFilter{EventID: eventId})
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	eventSize := len(entries)
	rankingUpdates := make(map[uuid.UUID]int32, eventSize)
	for _, entry := range entries {
		if entry.MembershipID != nil && entry.Placement > 0 {
			points := CalculatePoints(eventSize, int(entry.Placement), event.PointsMultiplier)
			rankingUpdates[*entry.MembershipID] = -int32(points)
		}
	}

	if err := tx.Rankings().BatchIncrementPoints(rankingUpdates); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *eventService) NewRebuy(eventId int32) error {
	event, err := svc.store.Events().FindByID(eventId)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return e.NotFound(err.Error())
		}
		return e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateEnded {
		return e.Forbidden("This event has already ended, it cannot be ended again.")
	}

	tx, err := svc.store.BeginTx()
	if err != nil {
		return e.InternalServerError(err.Error())
	}
	defer tx.Rollback()

	semester, err := tx.Semesters().FindByID(event.SemesterID)
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Semesters().IncrementBudget(event.SemesterID, float32(semester.RebuyFee)); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Events().Update(&event, map[string]any{"rebuys": event.Rebuys + 1}); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}
