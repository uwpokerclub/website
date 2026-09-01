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
	}, models.MembershipSourceAdmin)
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
	}, models.MembershipSourceAdmin)
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
	}, models.MembershipSourceAdmin)
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
	}, models.MembershipSourceAdmin)
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

func TestMembershipService_UpdateMembership_UnmarkPaid_ExecutiveFreeTrialFlagUntouched(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	semester.FreeTrialLimit = 1
	require.NoError(t, st.Semesters().Create(semester))

	event := &models.Event{Name: "Weekly", State: models.EventStateStarted}
	require.NoError(t, st.Events().Create(event))

	// Executive membership, currently paid, with attendance already at the limit-of-1. Unpaying
	// it must not trigger a free-trial recompute since executives are never subject to it.
	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: true, Executive: true}
	require.NoError(t, st.Memberships().Create(membership))
	require.NoError(t, st.Entries().Create(&models.Participant{MembershipID: &membership.ID, EventID: event.ID}))

	svc := NewMembershipService(st)
	paid := false
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.True(t, updated.FreeTrialAvailable, "unpaying an executive membership must not recompute the free-trial flag")

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.True(t, stored.FreeTrialAvailable)
}

func boolPtr(b bool) *bool {
	return &b
}

func TestMembershipService_CreateMembership_Executive(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	membership, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID:    1,
		Executive: true,
	}, models.MembershipSourceAdmin)
	require.NoError(t, err)
	require.True(t, membership.Executive)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(100), found.CurrentBudget, 0.001, "an executive membership must not increment the budget")
}

func TestMembershipService_CreateMembership_ExecutiveCannotBePaid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		req  *models.CreateMembershipRequest
	}{
		{
			name: "executive and paid",
			req:  &models.CreateMembershipRequest{UserID: 1, Executive: true, Paid: true},
		},
		{
			name: "executive and discounted",
			req:  &models.CreateMembershipRequest{UserID: 1, Executive: true, Paid: true, Discounted: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			st := inmemory.NewStore()
			semester := newTestSemesterForMembership()
			require.NoError(t, st.Semesters().Create(semester))

			svc := NewMembershipService(st)
			membership, err := svc.CreateMembership(semester.ID, tc.req, models.MembershipSourceAdmin)
			require.Nil(t, membership)
			require.ErrorIs(t, err, ErrExecutiveCannotBePaid)
		})
	}
}

func TestMembershipService_UpdateMembership_PaidToExecutive_ReversesFullFee(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: true}
	require.NoError(t, st.Memberships().Create(membership))

	before, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)

	svc := NewMembershipService(st)
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Executive: boolPtr(true)})
	require.NoError(t, err)
	require.True(t, updated.Executive)
	require.False(t, updated.Paid)
	require.False(t, updated.Discounted)

	after, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, before.CurrentBudget-float32(semester.MembershipFee), after.CurrentBudget, 0.001)

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.True(t, stored.Executive, "the write must actually persist, not just appear on the returned struct")
}

func TestMembershipService_UpdateMembership_DiscountedToExecutive_ReversesDiscountedFee(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: true, Discounted: true}
	require.NoError(t, st.Memberships().Create(membership))

	before, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)

	svc := NewMembershipService(st)
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Executive: boolPtr(true)})
	require.NoError(t, err)
	require.True(t, updated.Executive)
	require.False(t, updated.Paid)
	require.False(t, updated.Discounted)

	after, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, before.CurrentBudget-float32(semester.MembershipDiscountFee), after.CurrentBudget, 0.001)

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.True(t, stored.Executive, "the write must actually persist, not just appear on the returned struct")
}

func TestMembershipService_UpdateMembership_ExecutiveToNotExecutive_LeavesUnpaidNoBudgetCredit(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Executive: true}
	require.NoError(t, st.Memberships().Create(membership))

	before, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)

	svc := NewMembershipService(st)
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Executive: boolPtr(false)})
	require.NoError(t, err)
	require.False(t, updated.Executive)
	require.False(t, updated.Paid)

	after, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, before.CurrentBudget, after.CurrentBudget, 0.001)

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.False(t, stored.Executive, "the write must actually persist, not just appear on the returned struct")
}

func TestMembershipService_UpdateMembership_ExecutiveAndPaidInSameRequest_Rejected(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID}
	require.NoError(t, st.Memberships().Create(membership))

	before, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)

	svc := NewMembershipService(st)
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{
		Executive: boolPtr(true),
		Paid:      boolPtr(true),
	})
	require.Nil(t, updated)
	require.ErrorIs(t, err, ErrExecutiveCannotBePaid)

	after, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, before.CurrentBudget, after.CurrentBudget, 0.001, "a rejected update must not touch the budget")

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.False(t, stored.Executive)
	require.False(t, stored.Paid)
}

func TestMembershipService_UpdateMembership_PaidOnAlreadyExecutive_Rejected(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Executive: true}
	require.NoError(t, st.Memberships().Create(membership))

	before, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)

	svc := NewMembershipService(st)
	updated, err := svc.UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{
		Paid: boolPtr(true),
	})
	require.Nil(t, updated)
	require.ErrorIs(t, err, ErrExecutiveCannotBePaid, "setting paid:true on an already-executive membership must surface, not be silently discarded")

	after, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, before.CurrentBudget, after.CurrentBudget, 0.001, "a rejected update must not touch the budget")

	stored, err := st.Memberships().FindByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	require.True(t, stored.Executive)
	require.False(t, stored.Paid)
}

func TestMembershipService_CreateMembership_PersistsSource(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	membership, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID: 1,
		Paid:   true,
	}, models.MembershipSourceDiscord)

	require.NoError(t, err)
	require.NotNil(t, membership.Source)
	require.Equal(t, models.MembershipSourceDiscord, *membership.Source)
	require.NotNil(t, membership.CreatedAt, "in-memory Create must set CreatedAt - see Task 2")
}
