package services_test

import (
	"context"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type membershipPostgresLifecycleSpy struct {
	store.Store
	beginTxCalls  int
	commitCalls   int
	rollbackCalls int
}

func (s *membershipPostgresLifecycleSpy) BeginTx() (store.Store, error) {
	s.beginTxCalls++
	return s.Store.BeginTx()
}

func (s *membershipPostgresLifecycleSpy) Commit() error {
	s.commitCalls++
	return s.Store.Commit()
}

func (s *membershipPostgresLifecycleSpy) Rollback() error {
	s.rollbackCalls++
	return s.Store.Rollback()
}

func TestMembershipService_UsesCallerOwnedPostgresTransactions(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { container.Close(ctx) })
	db := container.GetDB()

	semester, err := testutils.CreateTestSemester(db, "Membership Transaction")
	require.NoError(t, err)
	rootStore := postgres.NewStore(db)

	// A caller-owned create keeps its row and paid budget credit invisible to
	// other connections until the caller commits.
	createUser := createMembershipTransactionUser(t, db)
	createTx, err := rootStore.BeginTx()
	require.NoError(t, err)
	createTxSpy := &membershipPostgresLifecycleSpy{Store: createTx}
	created, err := services.NewMembershipService(createTxSpy).CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID: createUser.ID,
		Paid:   true,
	}, models.MembershipSourceAdmin)
	require.NoError(t, err)
	require.NotNil(t, created)
	assertMembershipServiceDidNotManageTransaction(t, createTxSpy)

	inside, err := createTxSpy.Memberships().FindByID(created.ID)
	require.NoError(t, err)
	require.True(t, inside.Paid)
	insideSemester, err := createTxSpy.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(110), insideSemester.CurrentBudget, 0.001)
	assertMembershipExternallyAbsent(t, db, created.ID)
	assertMembershipBudget(t, db, semester.ID, 100)
	require.NoError(t, createTxSpy.Commit())
	require.Equal(t, 1, createTxSpy.commitCalls)
	require.Zero(t, createTxSpy.beginTxCalls)
	require.Zero(t, createTxSpy.rollbackCalls)
	assertMembershipExists(t, db, created.ID)
	assertMembershipBudget(t, db, semester.ID, 110)

	// A later composed write fails after membership creation and budget credit;
	// caller rollback must remove both effects.
	rollbackCreateUser := createMembershipTransactionUser(t, db)
	rollbackCreateTx, err := rootStore.BeginTx()
	require.NoError(t, err)
	rollbackCreateSpy := &membershipPostgresLifecycleSpy{Store: rollbackCreateTx}
	rolledBackCreate, err := services.NewMembershipService(rollbackCreateSpy).CreateMembership(semester.ID, &models.CreateMembershipRequest{
		UserID: rollbackCreateUser.ID,
		Paid:   true,
	}, models.MembershipSourceAdmin)
	require.NoError(t, err)
	insideSemester, err = rollbackCreateSpy.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(120), insideSemester.CurrentBudget, 0.001)
	duplicate := &models.Membership{UserID: rollbackCreateUser.ID, SemesterID: semester.ID}
	require.Error(t, rollbackCreateSpy.Memberships().Create(duplicate), "later composed write should fail on the user/semester unique constraint")
	assertMembershipServiceDidNotManageTransaction(t, rollbackCreateSpy)
	require.NoError(t, rollbackCreateSpy.Rollback())
	require.Equal(t, 1, rollbackCreateSpy.rollbackCalls)
	assertMembershipExternallyAbsent(t, db, rolledBackCreate.ID)
	assertMembershipBudget(t, db, semester.ID, 110)

	// The caller commits a paid transition and its fee delta.
	updateCommitMembership := createMembershipTransactionMembership(t, db, semester.ID)
	updateCommitTx, err := rootStore.BeginTx()
	require.NoError(t, err)
	updateCommitSpy := &membershipPostgresLifecycleSpy{Store: updateCommitTx}
	paid := true
	updated, err := services.NewMembershipService(updateCommitSpy).UpdateMembership(updateCommitMembership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.True(t, updated.Paid)
	assertMembershipServiceDidNotManageTransaction(t, updateCommitSpy)
	assertMembershipBudget(t, db, semester.ID, 110)
	require.NoError(t, updateCommitSpy.Commit())
	require.Equal(t, 1, updateCommitSpy.commitCalls)
	stored, err := postgres.NewStore(db).Memberships().FindByID(updateCommitMembership.ID)
	require.NoError(t, err)
	require.True(t, stored.Paid)
	assertMembershipBudget(t, db, semester.ID, 120)

	// A successful update can instead be rolled back by its caller.
	updateRollbackMembership := createMembershipTransactionMembership(t, db, semester.ID)
	updateRollbackTx, err := rootStore.BeginTx()
	require.NoError(t, err)
	updateRollbackSpy := &membershipPostgresLifecycleSpy{Store: updateRollbackTx}
	updated, err = services.NewMembershipService(updateRollbackSpy).UpdateMembership(updateRollbackMembership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.True(t, updated.Paid)
	insideSemester, err = updateRollbackSpy.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(130), insideSemester.CurrentBudget, 0.001)
	assertMembershipServiceDidNotManageTransaction(t, updateRollbackSpy)
	require.NoError(t, updateRollbackSpy.Rollback())
	stored, err = postgres.NewStore(db).Memberships().FindByID(updateRollbackMembership.ID)
	require.NoError(t, err)
	require.False(t, stored.Paid)
	assertMembershipBudget(t, db, semester.ID, 120)
}

func createMembershipTransactionUser(t *testing.T, db *gorm.DB) *models.User {
	t.Helper()
	unique := uuid.NewString()
	user, err := testutils.CreateTestUser(db, uint64(time.Now().UnixNano()), "Member", "Transaction", unique+"@example.com", models.FacultyMath, unique)
	require.NoError(t, err)
	return user
}

func createMembershipTransactionMembership(t *testing.T, db *gorm.DB, semesterID uuid.UUID) *models.Membership {
	t.Helper()
	user := createMembershipTransactionUser(t, db)
	membership := &models.Membership{UserID: user.ID, SemesterID: semesterID}
	require.NoError(t, db.Create(membership).Error)
	return membership
}

func assertMembershipServiceDidNotManageTransaction(t *testing.T, tx *membershipPostgresLifecycleSpy) {
	t.Helper()
	require.Zero(t, tx.beginTxCalls, "membership service must not begin a nested transaction")
	require.Zero(t, tx.commitCalls, "membership service must not commit the caller transaction")
	require.Zero(t, tx.rollbackCalls, "membership service must not roll back the caller transaction")
}

func assertMembershipExternallyAbsent(t *testing.T, db *gorm.DB, id uuid.UUID) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", id).Count(&count).Error)
	require.Zero(t, count)
}

func assertMembershipExists(t *testing.T, db *gorm.DB, id uuid.UUID) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", id).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func assertMembershipBudget(t *testing.T, db *gorm.DB, semesterID uuid.UUID, expected float32) {
	t.Helper()
	var budget float32
	require.NoError(t, db.Model(&models.Semester{}).Select("current_budget").Where("id = ?", semesterID).Scan(&budget).Error)
	require.InDelta(t, expected, budget, 0.001)
}
