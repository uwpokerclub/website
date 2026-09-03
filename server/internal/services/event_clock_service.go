package services

import (
	"errors"
	"time"

	"api/internal/models"
	"api/internal/store"
)

// ErrEmptyStructure is returned when an event's structure has no blind
// levels, so a clock cannot be derived for it.
var ErrEmptyStructure = errors.New("structure has no blind levels")

// ErrInvalidLevel is returned when SetLevel is given an index outside the
// structure's blind levels.
var ErrInvalidLevel = errors.New("level index out of range")

type eventClockService struct {
	store store.Store
}

func NewEventClockService(st store.Store) *eventClockService {
	return &eventClockService{
		store: st,
	}
}

// GetClock returns the fully derived, rolled-forward state of an event's
// clock. It lazily creates the clock row on first read but never writes the
// roll-forward itself: derivation is deterministic, so the stored row plus
// now always yields the truth.
func (s *eventClockService) GetClock(eventID int32) (models.DerivedClock, error) {
	levels, err := s.blindLevels(eventID)
	if err != nil {
		return models.DerivedClock{}, err
	}

	clock, err := s.store.EventClocks().FindByEventID(eventID)
	if errors.Is(err, store.ErrNotFound) {
		clock, err = s.lazilyCreate(eventID, levels, time.Now().UTC())
	}
	if err != nil {
		return models.DerivedClock{}, err
	}

	derived, ok := clock.Derive(levels, time.Now().UTC())
	if !ok {
		return models.DerivedClock{}, ErrEmptyStructure
	}

	return derived, nil
}

// Pause freezes the clock at its current, rolled-forward remaining time.
// Pausing an already-paused clock is a strict no-op: it does not persist and
// does not bump version.
func (s *eventClockService) Pause(eventID int32) (models.DerivedClock, error) {
	return s.applyAction(eventID, func(clock *models.EventClock, now time.Time) bool {
		if clock.PausedAt != nil {
			return false
		}
		clock.PausedAt = &now
		return true
	})
}

// Resume unfreezes the clock, restoring exactly the remaining time it had at
// the moment it was paused. Resuming an already-running clock is a strict
// no-op: it does not persist and does not bump version.
func (s *eventClockService) Resume(eventID int32) (models.DerivedClock, error) {
	return s.applyAction(eventID, func(clock *models.EventClock, now time.Time) bool {
		if clock.PausedAt == nil {
			return false
		}
		remaining := clock.LevelEndsAt.Sub(*clock.PausedAt)
		if remaining < 0 {
			remaining = 0
		}
		clock.LevelEndsAt = now.Add(remaining)
		clock.PausedAt = nil
		return true
	})
}

// Adjust shifts the current level's deadline by deltaSeconds, which may be
// negative. A negative adjustment that pushes the deadline past 0 rolls
// forward into the next level, carrying the overflow.
func (s *eventClockService) Adjust(eventID int32, deltaSeconds int) (models.DerivedClock, error) {
	return s.applyAction(eventID, func(clock *models.EventClock, now time.Time) bool {
		if deltaSeconds == 0 {
			return false
		}
		clock.LevelEndsAt = clock.LevelEndsAt.Add(time.Duration(deltaSeconds) * time.Second)
		return true
	})
}

// SetLevel jumps the clock to an absolute level index with a full, fresh
// duration for that level. It returns ErrInvalidLevel if index is outside
// the structure's blind levels.
func (s *eventClockService) SetLevel(eventID int32, index int32) (models.DerivedClock, error) {
	var derived models.DerivedClock
	var invalid error
	_, err := s.applyActionWithLevels(eventID, func(clock *models.EventClock, levels []time.Duration, now time.Time) bool {
		if index < 0 || int(index) >= len(levels) {
			invalid = ErrInvalidLevel
			return false
		}
		clock.LevelIndex = index
		clock.LevelEndsAt = now.Add(levels[index])
		if clock.PausedAt != nil {
			clock.PausedAt = &now
		}
		return true
	}, &derived)
	if err != nil {
		return models.DerivedClock{}, err
	}
	if invalid != nil {
		return models.DerivedClock{}, invalid
	}
	return derived, nil
}

func (s *eventClockService) blindLevels(eventID int32) ([]time.Duration, error) {
	event, err := s.store.Events().FindByID(eventID)
	if err != nil {
		return nil, err
	}

	// Events().FindByID preloads the structure (with blinds) on postgres, so
	// prefer that over a second round trip. The in-memory store does not
	// preload it, so fall back to a direct lookup there.
	structure := event.Structure
	if structure == nil {
		found, err := s.store.Structures().FindByID(event.StructureID)
		if err != nil {
			return nil, err
		}
		structure = &found
	}

	levels := make([]time.Duration, len(structure.Blinds))
	for i, blind := range structure.Blinds {
		levels[i] = time.Duration(blind.Time) * time.Minute
	}

	return levels, nil
}

// newInitialClock is the "paused, with a full level on the board" state a
// clock materialises into on first contact, whether that contact is a read
// or an action.
func newInitialClock(eventID int32, levels []time.Duration, now time.Time) models.EventClock {
	return models.EventClock{
		EventID:     eventID,
		LevelIndex:  0,
		LevelEndsAt: now.Add(levels[0]),
		PausedAt:    &now,
		Version:     1,
		UpdatedAt:   now,
	}
}

func (s *eventClockService) lazilyCreate(eventID int32, levels []time.Duration, now time.Time) (models.EventClock, error) {
	if len(levels) == 0 {
		return models.EventClock{}, ErrEmptyStructure
	}

	clock := newInitialClock(eventID, levels, now)

	if err := s.store.EventClocks().Create(&clock); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return s.store.EventClocks().FindByEventID(eventID)
		}
		return models.EventClock{}, err
	}

	return clock, nil
}

// applyAction is the shared derive-then-apply-then-persist path for actions
// that don't need the blind durations themselves, only the already-derived
// clock state.
func (s *eventClockService) applyAction(eventID int32, apply func(clock *models.EventClock, now time.Time) bool) (models.DerivedClock, error) {
	var derived models.DerivedClock
	_, err := s.applyActionWithLevels(eventID, func(clock *models.EventClock, levels []time.Duration, now time.Time) bool {
		return apply(clock, now)
	}, &derived)
	return derived, err
}

// applyActionWithLevels loads the event's blind durations, begins a
// transaction, locks the clock row (lazily creating it if this is the very
// first interaction with the clock), rolls it forward to now, lets apply
// mutate it, and persists the result with version bumped - unless apply
// reports no change, in which case nothing is written. The final derived
// state is written into out.
func (s *eventClockService) applyActionWithLevels(
	eventID int32,
	apply func(clock *models.EventClock, levels []time.Duration, now time.Time) bool,
	out *models.DerivedClock,
) (models.EventClock, error) {
	levels, err := s.blindLevels(eventID)
	if err != nil {
		return models.EventClock{}, err
	}
	if len(levels) == 0 {
		return models.EventClock{}, ErrEmptyStructure
	}

	tx, err := s.store.BeginTx()
	if err != nil {
		return models.EventClock{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	clock, err := tx.EventClocks().FindByEventIDForUpdate(eventID)
	wasCreated := false
	if errors.Is(err, store.ErrNotFound) {
		candidate := newInitialClock(eventID, levels, now)
		if err := tx.EventClocks().Create(&candidate); err != nil {
			if !errors.Is(err, store.ErrAlreadyExists) {
				return models.EventClock{}, err
			}
			// Lost the race to create the row within this transaction: a
			// concurrent transaction committed it first. Postgres' unique
			// index makes our INSERT block on the conflicting key until that
			// winner commits, so by the time ON CONFLICT DO NOTHING resolves
			// here, its row is already visible to us - fetch and lock it
			// like any other pre-existing clock, same as GetClock's
			// lazilyCreate does on the read path.
			clock, err = tx.EventClocks().FindByEventIDForUpdate(eventID)
			if err != nil {
				return models.EventClock{}, err
			}
		} else {
			wasCreated = true
			clock = candidate
		}
	} else if err != nil {
		return models.EventClock{}, err
	}

	derived, ok := clock.Derive(levels, now)
	if !ok {
		return models.EventClock{}, ErrEmptyStructure
	}
	clock.LevelIndex = derived.LevelIndex
	clock.LevelEndsAt = derived.LevelEndsAt
	clock.PausedAt = derived.PausedAt

	changed := apply(&clock, levels, now)

	// A strict no-op on a clock that already existed persists nothing at
	// all: pause-while-paused and resume-while-running must not bump
	// version. But if this transaction just lazily created the row, that
	// creation itself must be committed even if the requested action turned
	// out to be a no-op on top of it (e.g. pausing a clock that lazy
	// creation already started paused).
	if !changed && !wasCreated {
		*out = derived
		return clock, nil
	}

	if changed {
		clock.Version++
		clock.UpdatedAt = now

		if err := tx.EventClocks().Update(&clock); err != nil {
			return models.EventClock{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.EventClock{}, err
	}

	finalDerived, ok := clock.Derive(levels, now)
	if !ok {
		return models.EventClock{}, ErrEmptyStructure
	}
	*out = finalDerived

	return clock, nil
}
