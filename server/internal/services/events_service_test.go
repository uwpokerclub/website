package services

import (
	"api/internal/errors"
	"api/internal/models"
	"api/internal/store/inmemory"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventService_EndEvent(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	m1, m2 := uuid.New(), uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m1, EventID: event.ID}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &m2, EventID: event.ID}))

	svc := NewEventService(st)
	require.NoError(t, svc.EndEvent(event.ID))

	found, err := st.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.EqualValues(t, models.EventStateEnded, found.State)

	entries, _, err := st.Entries().List(&models.ListParticipantsFilter{EventID: event.ID})
	require.NoError(t, err)
	require.Len(t, entries, 2)

	placements := map[uint16]bool{}
	for _, entry := range entries {
		require.NotNil(t, entry.SignedOutAt)
		placements[entry.Placement] = true
	}
	require.Equal(t, map[uint16]bool{1: true, 2: true}, placements)
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
	event := &models.Event{Name: "Weekly", State: models.EventStateEnded, StartDate: time.Now().UTC(), PointsMultiplier: 1}
	require.NoError(t, st.Events().Create(event))

	membershipID := uuid.New()
	now := time.Now().UTC()
	require.NoError(t, st.Entries().Create(&models.Participant{
		MembershipID: &membershipID,
		EventID:      event.ID,
		Placement:    1,
		SignedOutAt:  &now,
	}))
	require.NoError(t, st.Rankings().Create(&models.Ranking{MembershipID: membershipID, Points: 10}))

	svc := NewEventService(st)
	require.NoError(t, svc.UndoEndEvent(event.ID))

	found, err := st.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.EqualValues(t, models.EventStateStarted, found.State)

	ranking, err := st.Rankings().FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.Less(t, ranking.Points, int32(10))
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
