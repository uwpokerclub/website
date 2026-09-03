package services

import (
	"api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventService_EndEvent_DistinctTimestamps(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	now := time.Now().UTC()
	t1, t2, t3 := now, now.Add(-time.Minute), now.Add(-2*time.Minute)
	m1, m2, m3 := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m1, EventID: event.ID, SignedOutAt: &t1}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m2, EventID: event.ID, SignedOutAt: &t2}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m3, EventID: event.ID, SignedOutAt: &t3}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))

	// N=3, no ties: place 1 -> 28, place 2 -> 11, place 3 -> 1 (25*ln(3/place)+1, rounded).
	want := map[uuid.UUID]int32{m1: 28, m2: 11, m3: 1}
	for membershipID, wantPoints := range want {
		entry, err := st.Entries().FindByMembershipAndEventID(membershipID, event.ID)
		require.NoError(t, err)
		require.EqualValues(t, wantPoints, entry.Points)

		ranking, err := st.Rankings().FindByMembershipID(membershipID)
		require.NoError(t, err)
		require.EqualValues(t, wantPoints, ranking.Points, "rankings.points must equal the entry's points")
	}
}

func TestEventService_EndEvent_TiedEntries(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	// Neither entry has signed out, so EndEvent's SignOutAllUnsigned call gives both the same
	// timestamp (event.StartDate) and they must be scored as one tie group, not by row order.
	m1, m2 := uuid.New(), uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m1, EventID: event.ID}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m2, EventID: event.ID}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))

	entries, _, err := st.Entries().List(&models.ListParticipantsFilter{EventID: event.ID})
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// mean(raw(2,1), raw(2,2)) rounded = 10 for both.
	for _, entry := range entries {
		require.NotNil(t, entry.SignedOutAt)
		require.EqualValues(t, 10, entry.Points)
	}
}

func TestEventService_EndEvent_EntryWithNoMembership(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	now := time.Now().UTC()
	earlier := now.Add(-time.Minute)
	m1 := uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: nil, EventID: event.ID, SignedOutAt: &now}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m1, EventID: event.ID, SignedOutAt: &earlier}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))

	entries, _, err := st.Entries().List(&models.ListParticipantsFilter{EventID: event.ID})
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// N=2, no ties: place 1 (no membership) -> 18, place 2 -> 1.
	for _, entry := range entries {
		if entry.MembershipID == nil {
			require.EqualValues(t, 18, entry.Points)
		} else {
			require.EqualValues(t, 1, entry.Points)
		}
	}
}

func TestEventService_EndEvent_AlreadyEnded(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateEnded, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventService(st)
	err := svc.EndEvent(event.ID)
	require.Error(t, err)
	apiErr, ok := err.(errors.APIErrorResponse)
	require.True(t, ok)
	require.Equal(t, 403, apiErr.Code)
}

func TestEventService_UndoEndEvent(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	m1, m2 := uuid.New(), uuid.New()
	require.NoError(t, st.Rankings().Create(&models.Ranking{MembershipID: m1, Points: 10}))
	require.NoError(t, st.Rankings().Create(&models.Ranking{MembershipID: m2, Points: 10}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m1, EventID: event.ID}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m2, EventID: event.ID}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))
	require.NoError(t, svc.UndoEndEvent(event.ID))

	found, err := st.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.EqualValues(t, models.EventStateStarted, found.State)

	for _, membershipID := range []uuid.UUID{m1, m2} {
		ranking, err := st.Rankings().FindByMembershipID(membershipID)
		require.NoError(t, err)
		require.EqualValues(t, 10, ranking.Points, "ranking must return to its exact pre-end value")
	}

	entries, _, err := st.Entries().List(&models.ListParticipantsFilter{EventID: event.ID})
	require.NoError(t, err)
	for _, entry := range entries {
		require.EqualValues(t, 0, entry.Points)
	}
}

func TestEventService_UndoEndEvent_ClearsClock(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))
	require.NoError(t, st.EventClocks().Create(&models.EventClock{
		EventID:     event.ID,
		LevelIndex:  3,
		LevelEndsAt: time.Now().UTC(),
		Version:     4,
		UpdatedAt:   time.Now().UTC(),
	}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))
	require.NoError(t, svc.UndoEndEvent(event.ID))

	_, err := st.EventClocks().FindByEventID(event.ID)
	require.ErrorIs(t, err, store.ErrNotFound, "restarting an event must clear its clock so it begins at level 1")
}

func TestEventService_UndoEndEvent_ClearsClock_NoClockIsNotAnError(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateEnded, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventService(st)
	require.NoError(t, svc.UndoEndEvent(event.ID))
}

func TestEventService_UndoEndEvent_ExactDespiteFieldChanges(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	now := time.Now().UTC()
	earlier := now.Add(-time.Minute)
	m1, m2 := uuid.New(), uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m1, EventID: event.ID, SignedOutAt: &now}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m2, EventID: event.ID, SignedOutAt: &earlier}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))

	// Simulate the entry composition changing between end and undo. A recompute-based undo would
	// use this new eventSize (3, not 2) and give the wrong answer; subtracting each entry's own
	// stored Points must not be affected by it at all.
	m3 := uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m3, EventID: event.ID}))

	require.NoError(t, svc.UndoEndEvent(event.ID))

	for _, membershipID := range []uuid.UUID{m1, m2} {
		ranking, err := st.Rankings().FindByMembershipID(membershipID)
		require.NoError(t, err)
		require.EqualValues(t, 0, ranking.Points)
	}
}

func TestEventService_UndoEndEvent_NotEnded(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventService(st)
	err := svc.UndoEndEvent(event.ID)
	require.Error(t, err)
	apiErr, ok := err.(errors.APIErrorResponse)
	require.True(t, ok)
	require.Equal(t, 403, apiErr.Code)
}

func TestEventService_NewRebuy(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := &models.Semester{Name: "Fall 2026", RebuyFee: 5, CurrentBudget: 100}
	require.NoError(t, st.Semesters().Create(semester))

	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), SemesterID: semester.ID, Rebuys: 0}
	require.NoError(t, st.Events().Create(event))

	svc := NewEventService(st)
	require.NoError(t, svc.NewRebuy(event.ID))

	foundEvent, err := st.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, foundEvent.Rebuys)

	foundSemester, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(105), foundSemester.CurrentBudget, 0.001)
}
