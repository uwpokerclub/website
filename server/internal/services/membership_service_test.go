package services

import (
	"testing"

	"api/internal/errors"
	"api/internal/models"
	"api/internal/store/inmemory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestSemesterForMembership() *models.Semester {
	return &models.Semester{
		Name:                  "Fall 2026",
		StartingBudget:        100,
		CurrentBudget:         100,
		MembershipFee:         10,
		MembershipDiscountFee: 5,
	}
}

func TestMembershipService_CreateMembership_InvalidState(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	_, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID:     1,
		Paid:       false,
		Discounted: true,
	})
	require.Error(t, err)
	_, isAPIError := err.(errors.APIErrorResponse)
	require.False(t, isAPIError)
	require.Equal(t, "cannot create membership that is not paid and discounted", err.Error())
}

func TestMembershipService_UpdateMembership_NotFound(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	paid := true
	membership, err := svc.UpdateMembership(uuid.New(), semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.Nil(t, membership)
}

func TestMembershipService_CreateMembership_Paid(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	_, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID: 1,
		Paid:   true,
	})
	require.NoError(t, err)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(110), found.CurrentBudget, 0.001)
}

func TestMembershipService_CreateMembership_PaidDiscounted(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	_, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID:     1,
		Paid:       true,
		Discounted: true,
	})
	require.NoError(t, err)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(105), found.CurrentBudget, 0.001)
}

func TestMembershipService_CreateMembership_Unpaid(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	_, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID: 1,
		Paid:   false,
	})
	require.NoError(t, err)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(100), found.CurrentBudget, 0.001)
}

func TestMembershipService_UpdateMembership_MarkPaid(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewMembershipService(st)
	paid := true
	_, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(110), found.CurrentBudget, 0.001)
}

// TestMembershipService_UpdateMembership_InvalidState covers the guard that
// rejects a membership left unpaid but discounted. Both cases reach it through
// a partial update, which is the only way to hit it now that the request DTO
// carries pointers: the final state is derived from the stored membership
// merged with whichever fields the caller actually sent.
func TestMembershipService_UpdateMembership_InvalidState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		paid       bool
		discounted bool
		req        *models.UpdateMembershipRequest
	}{
		{
			name:       "discounting an unpaid membership",
			paid:       false,
			discounted: false,
			req:        &models.UpdateMembershipRequest{Discounted: boolPtr(true)},
		},
		{
			name:       "unpaying a discounted membership",
			paid:       true,
			discounted: true,
			req:        &models.UpdateMembershipRequest{Paid: boolPtr(false)},
		},
		{
			name:       "both flipped at once",
			paid:       true,
			discounted: false,
			req:        &models.UpdateMembershipRequest{Paid: boolPtr(false), Discounted: boolPtr(true)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			st := inmemory.NewStore()
			semester := newTestSemesterForMembership()
			require.NoError(t, st.Semesters().Create(semester))

			membership := &models.Membership{
				UserID:     1,
				SemesterID: semester.ID,
				Paid:       tc.paid,
				Discounted: tc.discounted,
			}
			require.NoError(t, st.Memberships().Create(membership))

			before, err := st.Semesters().FindByID(semester.ID)
			require.NoError(t, err)

			svc := NewMembershipService(st)
			updated, err := svc.UpdateMembership(membership.ID, semester.ID, tc.req)

			require.Error(t, err)
			require.Nil(t, updated)
			_, isAPIError := err.(errors.APIErrorResponse)
			require.False(t, isAPIError)
			require.Equal(t, "cannot set membership to not paid and discounted", err.Error())

			// The guard runs before the transaction opens, so nothing should move.
			after, err := st.Semesters().FindByID(semester.ID)
			require.NoError(t, err)
			require.InDelta(t, before.CurrentBudget, after.CurrentBudget, 0.001)

			stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
			require.NoError(t, err)
			require.Equal(t, tc.paid, stored.Paid)
			require.Equal(t, tc.discounted, stored.Discounted)
		})
	}
}

func TestMembershipService_UpdateMembership_MarkPaid_ResetsStaleFreeTrialFlag(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	semester.FreeTrialLimit = 1
	require.NoError(t, st.Semesters().Create(semester))

	// Simulate a membership that already exhausted its free trial while unpaid.
	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))
	require.NoError(t, st.Memberships().SetFreeTrialAvailable(membership.ID, false))

	svc := NewMembershipService(st)
	paid := true
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.True(t, updated.FreeTrialAvailable, "marking a membership paid should reset a stale exhausted-trial flag")

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.True(t, stored.FreeTrialAvailable)
}

func TestMembershipService_UpdateMembership_UnmarkPaid_RecomputesFreeTrialFlag(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	semester.FreeTrialLimit = 1
	require.NoError(t, st.Semesters().Create(semester))

	event := &models.Event{Name: "Weekly", State: models.EventStateStarted}
	require.NoError(t, st.Events().Create(event))

	// Membership is currently paid, with attendance already at the limit-of-1 (irrelevant
	// while paid), and a stale "available" flag left over from before it was marked paid.
	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: true}
	require.NoError(t, st.Memberships().Create(membership))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: event.ID}))

	svc := NewMembershipService(st)
	paid := false
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.False(t, updated.FreeTrialAvailable, "unpaying a membership already at its attendance limit should recompute the flag as exhausted")

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.False(t, stored.FreeTrialAvailable)
}

func boolPtr(b bool) *bool {
	return &b
}
