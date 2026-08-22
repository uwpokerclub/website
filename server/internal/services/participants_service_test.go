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

func TestParticipantsService_CreateParticipant(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	svc := NewParticipantsService(st)
	membershipID := uuid.New()
	participant, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membershipID, EventID: event.ID})
	require.NoError(t, err)
	require.Equal(t, membershipID, *participant.MembershipID)
}

func TestParticipantsService_CreateParticipant_EventEnded(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateEnded, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	svc := NewParticipantsService(st)
	_, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: uuid.New(), EventID: event.ID})
	require.Error(t, err)
	apiErr, ok := err.(errors.APIErrorResponse)
	require.True(t, ok)
	require.Equal(t, 403, apiErr.Code)
}

func TestParticipantsService_UpdateParticipant_SignOut(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	membershipID := uuid.New()
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membershipID, EventID: event.ID}))

	svc := NewParticipantsService(st)
	participant, err := svc.UpdateParticipant(&models.UpdateParticipantRequest{MembershipID: membershipID, EventID: event.ID, SignOut: true})
	require.NoError(t, err)
	require.NotNil(t, participant.SignedOutAt)
}

func TestParticipantsService_UpdateParticipant_EventEnded(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event := &models.Event{Name: "Weekly", State: models.EventStateEnded, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	svc := NewParticipantsService(st)
	_, err := svc.UpdateParticipant(&models.UpdateParticipantRequest{MembershipID: uuid.New(), EventID: event.ID, SignOut: true})
	require.Error(t, err)
	apiErr, ok := err.(errors.APIErrorResponse)
	require.True(t, ok)
	require.Equal(t, 403, apiErr.Code)
}
