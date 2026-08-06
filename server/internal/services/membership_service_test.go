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

func TestMembershipService_CreateMembership_Paid(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	_, err := svc.CreateMembership(&models.CreateMembershipRequest{
		UserID:     1,
		SemesterID: semester.ID.String(),
		Paid:       true,
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
	_, err := svc.CreateMembership(&models.CreateMembershipRequest{
		UserID:     1,
		SemesterID: semester.ID.String(),
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
	_, err := svc.CreateMembership(&models.CreateMembershipRequest{
		UserID:     1,
		SemesterID: semester.ID.String(),
		Paid:       false,
	})
	require.NoError(t, err)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(100), found.CurrentBudget, 0.001)
}

func TestMembershipService_UpdateMembership_InvalidState(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewMembershipService(st)
	_, err := svc.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       false,
		Discounted: true,
	})
	require.Error(t, err)
	apiErr, ok := err.(errors.APIErrorResponse)
	require.True(t, ok)
	require.Equal(t, 400, apiErr.Code)
}

func TestMembershipService_UpdateMembership_MarkPaid(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	membership := &models.Membership{UserID: 1, SemesterID: semester.ID, Paid: false}
	require.NoError(t, st.Memberships().Create(membership))

	svc := NewMembershipService(st)
	_, err := svc.UpdateMembership(&models.UpdateMembershipRequest{
		ID:   membership.ID,
		Paid: true,
	})
	require.NoError(t, err)

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(110), found.CurrentBudget, 0.001)
}

func TestMembershipService_CreateMembershipV2_InvalidState(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	_, err := svc.CreateMembershipV2(semester.ID, &models.CreateMembershipRequestV2{
		UserID:     1,
		Paid:       false,
		Discounted: true,
	})
	require.Error(t, err)
	_, isAPIError := err.(errors.APIErrorResponse)
	require.False(t, isAPIError)
	require.Equal(t, "cannot create membership that is not paid and discounted", err.Error())
}

func TestMembershipService_UpdateMembershipV2_NotFound(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := newTestSemesterForMembership()
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewMembershipService(st)
	paid := true
	membership, err := svc.UpdateMembershipV2(uuid.New(), semester.ID, &models.UpdateMembershipRequestV2{Paid: &paid})
	require.NoError(t, err)
	require.Nil(t, membership)
}
