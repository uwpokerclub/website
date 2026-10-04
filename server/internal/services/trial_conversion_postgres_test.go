package services_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	apierrors "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

func TestParticipantsService_CreateParticipantSerializesFreeTrialLimitAtLimitMinusOne(t *testing.T) {
	container, semester, existingEvent, membership := newTrialConversionPostgresFixture(t)
	db := container.GetDB()
	var lockQueries atomic.Int64
	callbackName := "test:observe_membership_no_key_update"
	queryCallbacks := db.Callback().Query().After("gorm:query")
	require.NoError(t, queryCallbacks.Register(callbackName, func(tx *gorm.DB) {
		if strings.Contains(strings.ToUpper(tx.Statement.SQL.String()), "FOR NO KEY UPDATE") {
			lockQueries.Add(1)
		}
	}))
	t.Cleanup(func() { queryCallbacks.Remove(callbackName) })
	require.NoError(t, db.Model(&models.Semester{}).Where("id = ?", semester.ID).Update("free_trial_limit", 3).Error)

	structure, err := testutils.CreateTestStructure(db, "Concurrent trial entries")
	require.NoError(t, err)
	secondExistingEvent, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Already entered")
	require.NoError(t, err)
	firstRaceEvent, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Concurrent A")
	require.NoError(t, err)
	secondRaceEvent, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Concurrent B")
	require.NoError(t, err)
	for _, event := range []*models.Event{existingEvent, secondExistingEvent} {
		_, err := testutils.CreateTestParticipant(db, membership.ID, event.ID)
		require.NoError(t, err)
	}

	start := make(chan struct{})
	ready := make(chan struct{}, 2)
	results := make(chan error, 2)
	for _, event := range []*models.Event{firstRaceEvent, secondRaceEvent} {
		eventID := event.ID
		go func() {
			ready <- struct{}{}
			<-start
			_, err := services.NewParticipantsService(postgres.NewStore(db)).CreateParticipant(&models.CreateParticipantRequest{
				MembershipID: membership.ID,
				EventID:      eventID,
			})
			results <- err
		}()
	}
	<-ready
	<-ready
	close(start)

	successes, rejections := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			successes++
			continue
		}
		apiErr, ok := err.(apierrors.APIErrorResponse)
		require.True(t, ok, "unexpected participant error: %v", err)
		require.Equal(t, 403, apiErr.Code)
		rejections++
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, rejections)
	require.GreaterOrEqual(t, lockQueries.Load(), int64(2), "each service call must serialize through the explicit membership lock")

	count, err := postgres.NewStore(db).Entries().CountByMembershipID(membership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 3, count)
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

func TestMembershipService_UpdateMembershipRollbackKeepsConversionAndBudgetAtomic(t *testing.T) {
	container, semester, _, membership := newTrialConversionPostgresFixture(t)
	db := container.GetDB()
	startedAt := time.Date(2026, 9, 20, 18, 0, 0, 0, time.UTC)
	require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", membership.ID).Update("trial_started_at", startedAt).Error)

	// Fail the final paid update, after the service has incremented the semester
	// budget and stamped the first conversion. The transaction must roll both back.
	require.NoError(t, db.Exec(`
		CREATE FUNCTION reject_paid_membership_update() RETURNS trigger AS $$
		BEGIN
			IF NEW.paid AND NOT OLD.paid THEN
				RAISE EXCEPTION 'forced paid membership update failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER reject_paid_membership_update
		BEFORE UPDATE OF paid ON memberships
		FOR EACH ROW EXECUTE FUNCTION reject_paid_membership_update();
	`).Error)

	initialBudget := semester.CurrentBudget
	_, err := services.NewMembershipService(postgres.NewStore(db)).UpdateMembership(
		membership.ID,
		semester.ID,
		&models.UpdateMembershipRequest{Paid: trialBoolPtr(true)},
	)
	require.Error(t, err)

	stored, err := postgres.NewStore(db).Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, stored.Paid)
	require.NotNil(t, stored.TrialStartedAt)
	require.Equal(t, startedAt, *stored.TrialStartedAt)
	require.Nil(t, stored.ConvertedAt)
	var budgetAfter float32
	require.NoError(t, db.Model(&models.Semester{}).Select("current_budget").Where("id = ?", semester.ID).Scan(&budgetAfter).Error)
	require.Equal(t, initialBudget, budgetAfter)
}

func trialBoolPtr(value bool) *bool { return &value }
