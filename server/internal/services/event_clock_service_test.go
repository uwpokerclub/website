package services

import (
	"errors"
	"sync"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"

	"github.com/stretchr/testify/require"
)

// setupEventClock creates a structure with the given blind minute durations
// and an event referencing it, in a fresh in-memory store.
func setupEventClock(t *testing.T, blindMinutes ...int8) (st store.Store, eventID int32) {
	t.Helper()

	st = inmemory.NewStore()

	blinds := make([]models.Blind, len(blindMinutes))
	for i, m := range blindMinutes {
		blinds[i] = models.Blind{Index: int8(i), Small: 10, Big: 20, Time: m}
	}
	structure := &models.Structure{Name: "Test Structure", Blinds: blinds}
	require.NoError(t, st.Structures().Create(structure))

	event := &models.Event{Name: "Test Event", StructureID: structure.ID, State: models.EventStateStarted}
	require.NoError(t, st.Events().Create(event))

	return st, event.ID
}

func TestEventClockService_GetClock_PrefersPreloadedStructureOverExtraLookup(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()

	// A structure genuinely stored under the event's StructureID, but with
	// blinds deliberately different from what gets preloaded onto the event
	// itself - as postgres's Events().FindByID does via its Structure
	// preload. If blindLevels ever fell back to a fresh Structures().FindByID
	// lookup despite a preloaded Structure already being present, it would
	// pick up these wrong, 99-minute blinds instead.
	storedStructure := &models.Structure{Name: "Stored", Blinds: []models.Blind{{Index: 0, Time: 99}}}
	require.NoError(t, st.Structures().Create(storedStructure))

	preloadedStructure := &models.Structure{Blinds: []models.Blind{{Index: 0, Time: 5}}}
	event := &models.Event{
		Name:        "Test Event",
		StructureID: storedStructure.ID,
		Structure:   preloadedStructure,
		State:       models.EventStateStarted,
	}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventClockService(st)

	before := time.Now().UTC()
	derived, err := svc.GetClock(event.ID)
	after := time.Now().UTC()
	require.NoError(t, err)

	require.InDelta(t, 5*time.Minute, derived.Remaining, float64(after.Sub(before)),
		"GetClock must use the event's already-preloaded Structure rather than issuing a separate, redundant lookup")
}

func TestEventClockService_GetClock_LazyCreation(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	svc := NewEventClockService(st)

	before := time.Now().UTC()
	derived, err := svc.GetClock(eventID)
	after := time.Now().UTC()
	require.NoError(t, err)

	require.Equal(t, int32(0), derived.LevelIndex)
	require.NotNil(t, derived.PausedAt, "a freshly created clock must start paused")
	require.InDelta(t, 15*time.Minute, derived.Remaining, float64(after.Sub(before)))

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int32(0), stored.LevelIndex)
	require.NotNil(t, stored.PausedAt)
	require.Equal(t, int64(1), stored.Version)
}

func TestEventClockService_GetClock_IdempotentConcurrentFirstReads(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	svc := NewEventClockService(st)

	const n = 20
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = svc.GetClock(eventID)
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(1), stored.Version, "concurrent first reads must lazily create exactly one row")
}

func TestEventClockService_GetClock_ReadPathDoesNotWrite(t *testing.T) {
	t.Parallel()

	// Two five-minute levels; the stored clock is already past the first
	// level's end, so a read must roll forward across the boundary.
	st, eventID := setupEventClock(t, 5, 5)

	past := time.Now().UTC().Add(-10 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID:     eventID,
		LevelIndex:  0,
		LevelEndsAt: past,
		Version:     1,
		UpdatedAt:   past,
	}))

	svc := NewEventClockService(st)

	derived, err := svc.GetClock(eventID)
	require.NoError(t, err)
	require.Equal(t, int32(1), derived.LevelIndex, "the derived read must roll forward across the level boundary")

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int32(0), stored.LevelIndex, "the stored row must stay behind; reads never write")
	require.Equal(t, int64(1), stored.Version)
	require.WithinDuration(t, past, stored.UpdatedAt, 0)

	// A second read crossing the same boundary again must still not write.
	_, err = svc.GetClock(eventID)
	require.NoError(t, err)
	stored2, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, stored.Version, stored2.Version)
	require.WithinDuration(t, stored.UpdatedAt, stored2.UpdatedAt, 0)
}

func TestEventClockService_GetClock_EmptyStructure(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t)
	svc := NewEventClockService(st)

	_, err := svc.GetClock(eventID)
	require.True(t, errors.Is(err, ErrEmptyStructure))
}

func TestEventClockService_Pause_FromRunning(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	endsAt := time.Now().UTC().Add(10 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: nil, Version: 1, UpdatedAt: time.Now().UTC(),
	}))

	svc := NewEventClockService(st)

	derived, err := svc.Pause(eventID)
	require.NoError(t, err)
	require.NotNil(t, derived.PausedAt)
	require.InDelta(t, 0, derived.LevelEndsAt.Sub(endsAt), float64(time.Second))

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(2), stored.Version)
	require.NotNil(t, stored.PausedAt)
}

func TestEventClockService_Pause_NoopWhilePaused(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	pausedAt := time.Now().UTC().Add(-2 * time.Minute)
	endsAt := pausedAt.Add(10 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: &pausedAt, Version: 1, UpdatedAt: pausedAt,
	}))

	svc := NewEventClockService(st)

	_, err := svc.Pause(eventID)
	require.NoError(t, err)

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(1), stored.Version, "pausing an already-paused clock must not bump version")
	require.WithinDuration(t, pausedAt, *stored.PausedAt, 0, "re-pausing must not move paused_at forward and destroy remaining time")
}

func TestEventClockService_Resume_FromPaused(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	pausedAt := time.Now().UTC().Add(-2 * time.Minute)
	endsAt := pausedAt.Add(10 * time.Minute) // 10 minutes were remaining at the moment of pause
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: &pausedAt, Version: 1, UpdatedAt: pausedAt,
	}))

	svc := NewEventClockService(st)

	before := time.Now().UTC()
	derived, err := svc.Resume(eventID)
	after := time.Now().UTC()
	require.NoError(t, err)
	require.Nil(t, derived.PausedAt)
	require.InDelta(t, 10*time.Minute, derived.Remaining, float64(after.Sub(before)+time.Second))

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(2), stored.Version)
	require.Nil(t, stored.PausedAt)
}

func TestEventClockService_Resume_NoopWhileRunning(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	endsAt := time.Now().UTC().Add(10 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: nil, Version: 1, UpdatedAt: time.Now().UTC(),
	}))

	svc := NewEventClockService(st)

	_, err := svc.Resume(eventID)
	require.NoError(t, err)

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(1), stored.Version, "resuming an already-running clock must not bump version")
	require.WithinDuration(t, endsAt, stored.LevelEndsAt, 0, "resuming while running must not shift the deadline")
}

func TestEventClockService_Adjust_ShiftsDeadlineWhileRunning(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	endsAt := time.Now().UTC().Add(10 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: nil, Version: 1, UpdatedAt: time.Now().UTC(),
	}))

	svc := NewEventClockService(st)

	derived, err := svc.Adjust(eventID, 60)
	require.NoError(t, err)
	require.Equal(t, int32(0), derived.LevelIndex)
	require.InDelta(t, endsAt.Add(60*time.Second).Unix(), derived.LevelEndsAt.Unix(), 1)

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(2), stored.Version)
}

func TestEventClockService_Adjust_NegativeCrossesLevelBoundary(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 5, 10)
	now := time.Now().UTC()
	endsAt := now.Add(20 * time.Second) // 20s left on level 0
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: nil, Version: 1, UpdatedAt: now,
	}))

	svc := NewEventClockService(st)

	// -1 minute pushes 40s past the boundary into level 1.
	derived, err := svc.Adjust(eventID, -60)
	require.NoError(t, err)
	require.Equal(t, int32(1), derived.LevelIndex)
	require.InDelta(t, 10*time.Minute-40*time.Second, derived.Remaining, float64(2*time.Second))
}

func TestEventClockService_SetLevel(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 5, 10, 15)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: time.Now().UTC().Add(5 * time.Minute), PausedAt: nil, Version: 1, UpdatedAt: time.Now().UTC(),
	}))

	svc := NewEventClockService(st)

	before := time.Now().UTC()
	derived, err := svc.SetLevel(eventID, 2)
	after := time.Now().UTC()
	require.NoError(t, err)
	require.Equal(t, int32(2), derived.LevelIndex)
	require.InDelta(t, 15*time.Minute, derived.Remaining, float64(after.Sub(before)+time.Second))

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(2), stored.Version)
}

func TestEventClockService_SetLevel_OutOfRange(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 5, 10, 15)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: time.Now().UTC().Add(5 * time.Minute), PausedAt: nil, Version: 1, UpdatedAt: time.Now().UTC(),
	}))

	svc := NewEventClockService(st)

	_, err := svc.SetLevel(eventID, 3)
	require.True(t, errors.Is(err, ErrInvalidLevel))

	_, err = svc.SetLevel(eventID, -1)
	require.True(t, errors.Is(err, ErrInvalidLevel))

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(1), stored.Version, "a rejected out-of-range level must not persist")
}

func TestEventClockService_SetLevel_OutOfRange_FirstEverInteraction(t *testing.T) {
	t.Parallel()

	// No clock row exists yet: this SetLevel call is the very first contact
	// the event's clock has ever had, and the index is invalid.
	st, eventID := setupEventClock(t, 5, 10, 15)
	svc := NewEventClockService(st)

	_, err := svc.SetLevel(eventID, 5)
	require.True(t, errors.Is(err, ErrInvalidLevel))

	// Lazy creation still committed: the rejected level jump was layered on
	// top of a clock that had to be materialised to be evaluated at all.
	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int32(0), stored.LevelIndex)
	require.NotNil(t, stored.PausedAt)
	require.Equal(t, int64(1), stored.Version)
}

func TestEventClockService_Adjust_ZeroDeltaIsNoop(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	endsAt := time.Now().UTC().Add(10 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: endsAt, PausedAt: nil, Version: 1, UpdatedAt: time.Now().UTC(),
	}))

	svc := NewEventClockService(st)

	_, err := svc.Adjust(eventID, 0)
	require.NoError(t, err)

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(1), stored.Version, "adjusting by zero seconds must not bump version")
	require.WithinDuration(t, endsAt, stored.LevelEndsAt, 0)
}

func TestEventClockService_ActionAfterLongIdleGap_AppliesToRolledForwardLevel(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 5, 5, 5)
	past := time.Now().UTC().Add(-20 * time.Minute)
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID: eventID, LevelIndex: 0, LevelEndsAt: past.Add(5 * time.Minute), PausedAt: nil, Version: 1, UpdatedAt: past,
	}))

	svc := NewEventClockService(st)

	derived, err := svc.Pause(eventID)
	require.NoError(t, err)
	require.Equal(t, int32(2), derived.LevelIndex, "pause after an idle gap must apply to the rolled-forward level, not the stored one")
}

func TestEventClockService_Pause_FirstEverInteractionLazilyCreates(t *testing.T) {
	t.Parallel()

	st, eventID := setupEventClock(t, 15)
	svc := NewEventClockService(st)

	derived, err := svc.Pause(eventID)
	require.NoError(t, err)
	require.NotNil(t, derived.PausedAt, "pausing a clock that was already paused by lazy creation is a no-op")

	stored, err := st.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int64(1), stored.Version, "lazy creation plus a no-op pause must not bump past version 1")
}

// TestEventClockService_ControlActionsRejectEndedEvent guards against a TOCTOU
// race where a caller's own pre-check of event state (e.g. a controller
// reading the event before opening a transaction) goes stale between that
// read and the action's transaction. The service must re-check state itself,
// inside its own transaction, rather than trusting an outside caller.
func TestEventClockService_ControlActionsRejectEndedEvent(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	structure := &models.Structure{Name: "Test Structure", Blinds: []models.Blind{{Index: 0, Small: 10, Big: 20, Time: 15}}}
	require.NoError(t, st.Structures().Create(structure))
	event := &models.Event{Name: "Ended Event", StructureID: structure.ID, State: models.EventStateEnded}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventClockService(st)

	_, err := svc.Pause(event.ID)
	require.True(t, errors.Is(err, ErrEventEnded), "Pause must reject an ended event even without an outside pre-check")

	_, err = svc.Resume(event.ID)
	require.True(t, errors.Is(err, ErrEventEnded), "Resume must reject an ended event even without an outside pre-check")

	_, err = svc.Adjust(event.ID, 60)
	require.True(t, errors.Is(err, ErrEventEnded), "Adjust must reject an ended event even without an outside pre-check")

	_, err = svc.SetLevel(event.ID, 0)
	require.True(t, errors.Is(err, ErrEventEnded), "SetLevel must reject an ended event even without an outside pre-check")

	_, err = st.EventClocks().FindByEventID(event.ID)
	require.True(t, errors.Is(err, store.ErrNotFound), "a rejected control action on an ended event must not lazily create a clock row")
}

func TestEventClockService_GetClock_AllowedOnEndedEvent(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	structure := &models.Structure{Name: "Test Structure", Blinds: []models.Blind{{Index: 0, Small: 10, Big: 20, Time: 15}}}
	require.NoError(t, st.Structures().Create(structure))
	event := &models.Event{Name: "Ended Event", StructureID: structure.ID, State: models.EventStateEnded}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventClockService(st)

	_, err := svc.GetClock(event.ID)
	require.NoError(t, err, "reading the clock of an ended event must remain allowed")
}
