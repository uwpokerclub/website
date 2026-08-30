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

	semester := &models.Semester{Name: "Fall 2026"}
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: true}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewParticipantsService(st)
	participant, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event.ID})
	require.NoError(t, err)
	require.Equal(t, membership.ID, *participant.MembershipID)
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

func TestParticipantsService_CreateParticipant_UnpaidNoFreeTrialsLeft(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event1 := &models.Event{Name: "Weekly 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event1))
	event2 := &models.Event{Name: "Weekly 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event2))
	event3 := &models.Event{Name: "Weekly 3", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event3))

	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 2}
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	// Attendance already at the limit before this call — the two prior entries are created
	// directly against the store, bypassing the service, to set up real attendance rather than
	// trusting a manually-set flag (the gate recomputes from attendance, not the flag).
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: event1.ID}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: event2.ID}))

	svc := NewParticipantsService(st)
	_, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event3.ID})
	require.Error(t, err)
	apiErr, ok := err.(errors.APIErrorResponse)
	require.True(t, ok)
	require.Equal(t, 403, apiErr.Code)
}

func TestParticipantsService_CreateParticipant_UnpaidExhaustsFreeTrial(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event1 := &models.Event{Name: "Weekly 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event1))
	event2 := &models.Event{Name: "Weekly 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event2))

	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 2}
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewParticipantsService(st)

	_, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event1.ID})
	require.NoError(t, err)

	updated, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.True(t, updated.FreeTrialAvailable, "should still have a free trial after 1 of 2 events")

	_, err = svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event2.ID})
	require.NoError(t, err)

	updated, err = st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, updated.FreeTrialAvailable, "free trial should be exhausted after the 2nd of 2 events")
}

func TestParticipantsService_CreateParticipant_PaidBypassesFreeTrialCheck(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event1 := &models.Event{Name: "Weekly 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event1))
	event2 := &models.Event{Name: "Weekly 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event2))

	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 1}
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: true}
	require.NoError(t, st.Memberships().Create(membership))
	// Attendance already at (in fact over) the limit, and the cached flag already says
	// exhausted — a paid membership must bypass both.
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: event1.ID}))
	require.NoError(t, st.Memberships().SetFreeTrialAvailable(membership.ID, false))

	svc := NewParticipantsService(st)
	participant, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event2.ID})
	require.NoError(t, err)
	require.Equal(t, membership.ID, *participant.MembershipID)
}

func TestParticipantsService_CreateParticipant_UnpaidNoLimitConfigured(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event1 := &models.Event{Name: "Weekly 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event1))
	event2 := &models.Event{Name: "Weekly 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event2))

	// FreeTrialLimit left at its zero value: the free-trial check must be a no-op, not "zero
	// free events allowed."
	semester := &models.Semester{Name: "Fall 2026"}
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewParticipantsService(st)

	_, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event1.ID})
	require.NoError(t, err)
	_, err = svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event2.ID})
	require.NoError(t, err)

	updated, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.True(t, updated.FreeTrialAvailable, "flag should be untouched when no limit is configured")
}

func TestParticipantsService_CreateParticipant_MidTermLimitIncreaseSelfHeals(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	event1 := &models.Event{Name: "Weekly 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event1))
	event2 := &models.Event{Name: "Weekly 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event2))

	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 1}
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewParticipantsService(st)

	// First event exhausts the limit-of-1 and flips the flag false.
	_, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event1.ID})
	require.NoError(t, err)

	blocked, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, blocked.FreeTrialAvailable)

	// Still blocked under the old limit.
	_, err = svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event2.ID})
	require.Error(t, err)

	// The exec raises the limit mid-term (1 -> 3), matching the issue's own motivation.
	semester.FreeTrialLimit = 3
	require.NoError(t, st.Semesters().Update(semester))

	// Now succeeds, and the previously-stale flag self-heals back to true, because the block
	// (and the sync) both recompute live rather than trusting the cached value.
	_, err = svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: event2.ID})
	require.NoError(t, err)

	healed, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.True(t, healed.FreeTrialAvailable, "raising the limit mid-term should un-block and re-flip the cached flag")
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

func TestParticipantsService_DeleteParticipant_RestoresFreeTrial(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 2}
	require.NoError(t, st.Semesters().Create(semester))

	first := &models.Event{Name: "Week 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(first))
	second := &models.Event{Name: "Week 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(second))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewParticipantsService(st)
	_, err := svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: first.ID})
	require.NoError(t, err)
	_, err = svc.CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: second.ID})
	require.NoError(t, err)

	exhausted, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, exhausted.FreeTrialAvailable, "two entries against a limit of two exhausts the trial")

	require.NoError(t, svc.DeleteParticipant(membership.ID, second.ID))

	restored, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.True(t, restored.FreeTrialAvailable, "deleting an entry drops attendance below the limit, so the trial is available again")
}

func TestParticipantsService_DeleteParticipant_StaysExhaustedAboveLimit(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 1}
	require.NoError(t, st.Semesters().Create(semester))

	first := &models.Event{Name: "Week 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(first))
	second := &models.Event{Name: "Week 2", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(second))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	// Two entries against a limit of one; the second bypasses the service to model entries
	// that predate the limit being lowered.
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: first.ID}))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: second.ID}))
	require.NoError(t, st.Memberships().SetFreeTrialAvailable(membership.ID, false))

	svc := NewParticipantsService(st)
	require.NoError(t, svc.DeleteParticipant(membership.ID, second.ID))

	stored, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, stored.FreeTrialAvailable, "one remaining entry still meets the limit of one")
}

func TestParticipantsService_DeleteParticipant_PaidMembershipUntouched(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := &models.Semester{Name: "Fall 2026", FreeTrialLimit: 1}
	require.NoError(t, st.Semesters().Create(semester))

	event := &models.Event{Name: "Week 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: true}
	require.NoError(t, st.Memberships().Create(membership))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: event.ID}))
	require.NoError(t, st.Memberships().SetFreeTrialAvailable(membership.ID, false))

	svc := NewParticipantsService(st)
	require.NoError(t, svc.DeleteParticipant(membership.ID, event.ID))

	stored, err := st.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, stored.FreeTrialAvailable, "a paid membership's flag is never recomputed")
}

func TestParticipantsService_DeleteParticipant_NotFound(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := &models.Semester{Name: "Fall 2026"}
	require.NoError(t, st.Semesters().Create(semester))

	event := &models.Event{Name: "Week 1", State: models.EventStateStarted, StartDate: time.Now().UTC()}
	require.NoError(t, st.Events().Create(event))

	membership := &models.Membership{SemesterID: semester.ID, UserID: 1, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewParticipantsService(st)
	require.ErrorIs(t, svc.DeleteParticipant(membership.ID, event.ID), ErrEntryNotFound)
	require.ErrorIs(t, svc.DeleteParticipant(uuid.New(), event.ID), ErrMembershipNotFound)
}
