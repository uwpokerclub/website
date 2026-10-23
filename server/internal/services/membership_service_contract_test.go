package services_test

import (
	"context"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"api/internal/store/inmemory"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMembershipStoreContract(t *testing.T) {
	t.Run("in-memory", func(t *testing.T) {
		st := inmemory.NewStore()
		runMembershipStoreContract(t, st, 411, false)
	})

	t.Run("postgres", func(t *testing.T) {
		ctx := context.Background()
		container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
		require.NoError(t, err)
		defer container.Close(ctx)

		db := container.GetDB()
		userID := uint64(time.Now().UnixNano())
		_, err = testutils.CreateTestUser(db, userID, "Member", "Contract", uuid.NewString()+"@example.com", "Math", uuid.NewString())
		require.NoError(t, err)
		runMembershipStoreContract(t, postgres.NewStore(db), userID, true)
		testPostgresMembershipConflictKeepsTransactionUsable(t, db, userID)
		testPostgresConcurrentExecutivePromotionRefundsOnce(t, db, userID)
	})
}

func testPostgresMembershipConflictKeepsTransactionUsable(t *testing.T, db *gorm.DB, userID uint64) {
	t.Helper()
	semester, err := testutils.CreateTestSemester(db, "Concurrent duplicate")
	require.NoError(t, err)
	root := postgres.NewStore(db)
	tx1, err := root.BeginTx()
	require.NoError(t, err)
	tx2, err := root.BeginTx()
	require.NoError(t, err)

	first := &models.Membership{UserID: userID, SemesterID: semester.ID}
	require.NoError(t, tx1.Memberships().Create(first))
	second := &models.Membership{UserID: userID, SemesterID: semester.ID}
	started := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		close(started)
		finished <- tx2.Memberships().Create(second)
	}()
	<-started
	require.NoError(t, tx1.Commit())
	require.ErrorIs(t, <-finished, store.ErrAlreadyExists)

	// A PostgreSQL uniqueness error would abort tx2 here. The targeted conflict
	// insert must leave it available for both a read and a write before commit.
	found, err := tx2.Memberships().FindByUserAndSemester(userID, semester.ID)
	require.NoError(t, err)
	require.Equal(t, first.ID, found.ID)
	require.NoError(t, tx2.Memberships().SetFreeTrialAvailable(found.ID, false))
	require.NoError(t, tx2.Commit())

	stored, err := root.Memberships().FindByUserAndSemester(userID, semester.ID)
	require.NoError(t, err)
	require.False(t, stored.FreeTrialAvailable)
}

func testPostgresConcurrentExecutivePromotionRefundsOnce(t *testing.T, db *gorm.DB, userID uint64) {
	t.Helper()
	semester, err := testutils.CreateTestSemester(db, "Concurrent promotion")
	require.NoError(t, err)
	membership := &models.Membership{UserID: userID, SemesterID: semester.ID, Paid: true}
	require.NoError(t, db.Create(membership).Error)
	require.NoError(t, db.Model(semester).Update("current_budget", 110).Error)

	root := postgres.NewStore(db)
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			tx, err := root.BeginTx()
			if err != nil {
				results <- err
				return
			}
			<-start
			_, err = services.NewMembershipService(tx).EnsureExecutiveMembership(userID, semester.ID, models.MembershipSourceAdmin)
			if err == nil {
				err = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
			results <- err
		}()
	}
	close(start)
	require.NoError(t, <-results)
	require.NoError(t, <-results)

	updated, err := root.Memberships().FindByUserAndSemester(userID, semester.ID)
	require.NoError(t, err)
	require.True(t, updated.Executive)
	updatedSemester, err := root.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, 100, updatedSemester.CurrentBudget, 0.001, "concurrent promotions must refund the paid fee only once")
}

func runMembershipStoreContract(t *testing.T, st store.Store, userID uint64, hasAssociations bool) {
	t.Helper()
	repo := st.Memberships()
	newSemester := func(name string, budget float32) *models.Semester {
		t.Helper()
		semester := &models.Semester{
			ID:                    uuid.New(),
			Name:                  name,
			StartingBudget:        100,
			CurrentBudget:         budget,
			MembershipFee:         10,
			MembershipDiscountFee: 5,
		}
		require.NoError(t, st.Semesters().Create(semester))
		return semester
	}

	missingSemester := newSemester("lookup missing", 100)
	_, err := repo.FindByUserAndSemester(userID, missingSemester.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	otherSemester := newSemester("other semester", 110)
	other := &models.Membership{UserID: userID, SemesterID: otherSemester.ID, Paid: true}
	require.NoError(t, repo.Create(other))
	_, err = repo.FindByUserAndSemester(userID, missingSemester.ID)
	require.ErrorIs(t, err, store.ErrNotFound, "lookup must scope by semester as well as user")
	foundOther, err := repo.FindByUserAndSemester(userID, otherSemester.ID)
	require.NoError(t, err)
	require.Equal(t, other.ID, foundOther.ID)
	if hasAssociations {
		require.NotNil(t, foundOther.User)
		require.NotNil(t, foundOther.Semester)
	}

	duplicate := &models.Membership{UserID: userID, SemesterID: otherSemester.ID}
	require.ErrorIs(t, repo.Create(duplicate), store.ErrAlreadyExists)

	createdSemester := newSemester("create executive", 100)
	created, err := services.NewMembershipService(st).EnsureExecutiveMembership(userID, createdSemester.ID, models.MembershipSourceDiscord)
	require.NoError(t, err)
	require.True(t, created.Executive)
	require.False(t, created.Paid)
	require.False(t, created.Discounted)
	require.NotEqual(t, uuid.Nil, created.ID)
	require.NotNil(t, created.Source)
	require.Equal(t, models.MembershipSourceDiscord, *created.Source)
	if hasAssociations {
		require.NotNil(t, created.User)
		require.NotNil(t, created.Semester)
	}
	budget, err := st.Semesters().FindByID(createdSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 100, budget.CurrentBudget, 0.001)

	paidSemester := newSemester("promote paid", 110)
	createdAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	trialStartedAt := createdAt.Add(time.Hour)
	convertedAt := createdAt.Add(2 * time.Hour)
	source := models.MembershipSourceDiscord
	paid := &models.Membership{
		UserID:             userID,
		SemesterID:         paidSemester.ID,
		Paid:               true,
		FreeTrialAvailable: false,
		CreatedAt:          &createdAt,
		Source:             &source,
		TrialStartedAt:     &trialStartedAt,
		ConvertedAt:        &convertedAt,
	}
	require.NoError(t, repo.Create(paid))
	require.NoError(t, repo.SetFreeTrialAvailable(paid.ID, false))
	service := services.NewMembershipService(st)
	result, err := service.EnsureExecutiveMembership(userID, paidSemester.ID, models.MembershipSourceAdmin)
	require.NoError(t, err)
	require.Equal(t, paid.ID, result.ID)
	require.True(t, result.Executive)
	require.False(t, result.Paid)
	require.False(t, result.Discounted)
	require.Equal(t, source, *result.Source, "promotion must preserve the original source")
	require.Equal(t, createdAt, *result.CreatedAt)
	require.Equal(t, trialStartedAt, *result.TrialStartedAt)
	require.Equal(t, convertedAt, *result.ConvertedAt)
	require.False(t, result.FreeTrialAvailable)
	budget, err = st.Semesters().FindByID(paidSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 100, budget.CurrentBudget, 0.001, "paid fee should be reversed exactly once")
	_, err = service.EnsureExecutiveMembership(userID, paidSemester.ID, models.MembershipSourceAdmin)
	require.NoError(t, err)
	budget, err = st.Semesters().FindByID(paidSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 100, budget.CurrentBudget, 0.001, "repeated upsert must be idempotent")

	discountedSemester := newSemester("promote discounted", 105)
	discounted := &models.Membership{UserID: userID, SemesterID: discountedSemester.ID, Paid: true, Discounted: true}
	require.NoError(t, repo.Create(discounted))
	_, err = service.EnsureExecutiveMembership(userID, discountedSemester.ID, models.MembershipSourceAdmin)
	require.NoError(t, err)
	budget, err = st.Semesters().FindByID(discountedSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 100, budget.CurrentBudget, 0.001, "discounted fee should be reversed exactly once")

	stillPaid, err := repo.FindByUserAndSemester(userID, otherSemester.ID)
	require.NoError(t, err)
	require.Equal(t, other.ID, stillPaid.ID, "same user membership in another semester must remain untouched")
	require.True(t, stillPaid.Paid)
	budget, err = st.Semesters().FindByID(otherSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 110, budget.CurrentBudget, 0.001)

	rollbackSemester := newSemester("rollback ensure", 110)
	rollbackMembership := &models.Membership{UserID: userID, SemesterID: rollbackSemester.ID, Paid: true}
	require.NoError(t, repo.Create(rollbackMembership))
	tx, err := st.BeginTx()
	require.NoError(t, err)
	_, err = services.NewMembershipService(tx).EnsureExecutiveMembership(userID, rollbackSemester.ID, models.MembershipSourceAdmin)
	require.NoError(t, err)
	txBudget, err := tx.Semesters().FindByID(rollbackSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 100, txBudget.CurrentBudget, 0.001)
	require.NoError(t, tx.Rollback())
	rollbackResult, err := repo.FindByUserAndSemester(userID, rollbackSemester.ID)
	require.NoError(t, err)
	require.Equal(t, rollbackMembership.ID, rollbackResult.ID)
	require.True(t, rollbackResult.Paid)
	require.False(t, rollbackResult.Executive)
	budget, err = st.Semesters().FindByID(rollbackSemester.ID)
	require.NoError(t, err)
	require.InDelta(t, 110, budget.CurrentBudget, 0.001)

	// Explicit IDs are a separate unique constraint and must not be disguised as
	// a user-semester duplicate.
	primaryKeyDuplicate := &models.Membership{ID: other.ID, UserID: userID, SemesterID: missingSemester.ID}
	err = repo.Create(primaryKeyDuplicate)
	require.Error(t, err)
	require.NotErrorIs(t, err, store.ErrAlreadyExists)
}
