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

	// Every entry now has a non-nil SignedOutAt, so List's "signed_out_at DESC" order is the
	// placement order. Consecutive entries sharing a SignedOutAt - most often the bulk block
	// SignOutAllUnsigned just created - are scored as one tie group, so points don't depend on
	// their arbitrary order within it.
	//
	// Note anyone still seated at force-end places last, not first: SignOutAllUnsigned stamps
	// them with event.StartDate. That predates this change, but the curve widens its cost.
	entries, _, err := tx.Entries().List(&models.ListParticipantsFilter{EventID: eventId})
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	eventSize := len(entries)
	rankingUpdates := make(map[uuid.UUID]int32, eventSize)
	pointsUpdates := make(map[int32]int32, eventSize)

	for i := 0; i < len(entries); {
		j := i
		for j+1 < len(entries) && entries[j+1].SignedOutAt != nil && entries[i].SignedOutAt != nil &&
			entries[j+1].SignedOutAt.Equal(*entries[i].SignedOutAt) {
			j++
		}

		groupPoints := int32(CalculateTiePoints(eventSize, i+1, j+1, event.PointsMultiplier))
		for k := i; k <= j; k++ {
			entry := entries[k]
			pointsUpdates[entry.ID] = groupPoints
			if entry.MembershipID != nil {
				rankingUpdates[*entry.MembershipID] += groupPoints
			}
		}

		i = j + 1
	}

	if err := tx.Entries().BatchUpdatePoints(pointsUpdates); err != nil {
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

	// Subtract each entry's stored Points rather than recomputing: eventSize and tie grouping
	// can differ at undo time, so a recompute would not reverse what EndEvent applied.
	rankingUpdates := make(map[uuid.UUID]int32, len(entries))
	pointsReset := make(map[int32]int32, len(entries))
	for _, entry := range entries {
		if entry.MembershipID != nil && entry.Points != 0 {
			rankingUpdates[*entry.MembershipID] -= entry.Points
		}
		pointsReset[entry.ID] = 0
	}

	if err := tx.Rankings().BatchIncrementPoints(rankingUpdates); err != nil {
		return e.InternalServerError(err.Error())
	}

	if err := tx.Entries().BatchUpdatePoints(pointsReset); err != nil {
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
