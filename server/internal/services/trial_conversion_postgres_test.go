package services_test

import (
	"context"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/stretchr/testify/require"
)

func newTrialConversionPostgresFixture(t *testing.T) (*testutils.PostgresTestContainer, *models.Semester, *models.Event, *models.Membership) {
	t.Helper()
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { container.Close(ctx) })
	db := container.GetDB()
	semester, err := testutils.CreateTestSemester(db, "Trial Conversion")
	require.NoError(t, err)
	require.NoError(t, db.Model(&models.Semester{}).Where("id = ?", semester.ID).Update("free_trial_limit", 4).Error)
	structure, err := testutils.CreateTestStructure(db, "Trial Conversion")
	require.NoError(t, err)
	event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Trial Conversion")
	require.NoError(t, err)
	userID := uint64(time.Now().UnixNano())
	user, err := testutils.CreateTestUser(db, userID, "Trial", "Player", "trial-player@example.com", models.FacultyMath, "trial-player")
	require.NoError(t, err)
	membership := &models.Membership{UserID: user.ID, SemesterID: semester.ID, Paid: false}
	require.NoError(t, db.Create(membership).Error)
	return container, semester, event, membership
}

func TestParticipantsService_CreateParticipantSerializesWithPaidTransition(t *testing.T) {
	container, semester, event, membership := newTrialConversionPostgresFixture(t)
	db := container.GetDB()
	baseStore := postgres.NewStore(db)
	lockingTx, err := baseStore.BeginTx()
	require.NoError(t, err)
	locked, err := lockingTx.Memberships().LockByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)

	type entryResult struct{ err error }
	result := make(chan entryResult, 1)
	go func() {
		_, err := services.NewParticipantsService(baseStore).CreateParticipant(&models.CreateParticipantRequest{
			MembershipID: membership.ID,
			EventID:      event.ID,
		})
		result <- entryResult{err: err}
	}()

	select {
	case result := <-result:
		t.Fatalf("entry creation must wait for the membership lock (returned: %v)", result.err)
	case <-time.After(200 * time.Millisecond):
	}

	locked.Paid = true
	require.NoError(t, lockingTx.Memberships().Update(&locked))
	require.NoError(t, lockingTx.Commit())

	select {
	case result := <-result:
		require.NoError(t, result.err)
	case <-time.After(5 * time.Second):
		t.Fatal("entry creation did not resume after the paid transition committed")
	}

	stored, err := baseStore.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.True(t, stored.Paid)
	require.Nil(t, stored.TrialStartedAt, "entry must reread paid status after acquiring the lock")
	require.Nil(t, stored.ConvertedAt)
}

func TestMembershipService_UpdateMembershipSerializesAfterTrialStartAndPreservesFirstConversion(t *testing.T) {
	container, semester, _, membership := newTrialConversionPostgresFixture(t)
	db := container.GetDB()
	baseStore := postgres.NewStore(db)
	lockingTx, err := baseStore.BeginTx()
	require.NoError(t, err)
	_, err = lockingTx.Memberships().LockByIDAndSemesterID(membership.ID, semester.ID)
	require.NoError(t, err)
	startedAt := time.Date(2026, 9, 20, 18, 0, 0, 0, time.UTC)
	require.NoError(t, lockingTx.Memberships().SetTrialStartedAtIfNull(membership.ID, startedAt))

	type updateResult struct {
		membership *models.Membership
		err        error
	}
	result := make(chan updateResult, 1)
	go func() {
		updated, err := services.NewMembershipService(baseStore).UpdateMembership(
			membership.ID,
			semester.ID,
			&models.UpdateMembershipRequest{Paid: trialBoolPtr(true), Discounted: trialBoolPtr(true)},
		)
		result <- updateResult{membership: updated, err: err}
	}()

	select {
	case result := <-result:
		t.Fatalf("membership update must wait for the trial-start lock (returned: %v)", result.err)
	case <-time.After(200 * time.Millisecond):
	}

	require.NoError(t, lockingTx.Commit())
	var updated *models.Membership
	select {
	case result := <-result:
		require.NoError(t, result.err)
		updated = result.membership
	case <-time.After(5 * time.Second):
		t.Fatal("membership update did not resume after the trial start committed")
	}
	require.True(t, updated.Paid)
	require.True(t, updated.Discounted)
	require.Equal(t, startedAt, *updated.TrialStartedAt)
	require.NotNil(t, updated.ConvertedAt)
	stored, err := baseStore.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	firstConvertedAt := *stored.ConvertedAt

	_, err = services.NewMembershipService(baseStore).UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: trialBoolPtr(false), Discounted: trialBoolPtr(false)})
	require.NoError(t, err)
	updated, err = services.NewMembershipService(baseStore).UpdateMembership(membership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: trialBoolPtr(true)})
	require.NoError(t, err)
	require.Equal(t, firstConvertedAt, *updated.ConvertedAt, "paid reversals must not replace the first conversion timestamp")
}

func trialBoolPtr(value bool) *bool { return &value }
