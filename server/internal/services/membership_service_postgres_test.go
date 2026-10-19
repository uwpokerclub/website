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

func TestMembershipService_InTxMethodsComposeWithCallerOwnedPostgresTransactions(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { container.Close(ctx) })
	db := container.GetDB()

	semester, err := testutils.CreateTestSemester(db, "Membership Transaction")
	require.NoError(t, err)
	rootSpy := &membershipPostgresLifecycleSpy{Store: postgres.NewStore(db)}
	svc := services.NewMembershipService(rootSpy)

	// A caller-owned create keeps both the membership and paid budget credit
	// invisible outside the transaction until the caller commits.
	createUser := createMembershipTransactionUser(t, db)
	createTx, err := rootSpy.BeginTx()
	require.NoError(t, err)
	createTxSpy := &membershipPostgresLifecycleSpy{Store: createTx}
	created, err := svc.CreateMembershipInTx(createTxSpy, semester.ID, &models.CreateMembershipRequest{
		UserID: createUser.ID,
		Paid:   true,
	}, models.MembershipSourceAdmin)
	require.NoError(t, err)
	require.NotNil(t, created)
	assertMembershipTxLifecycleUntouched(t, rootSpy, createTxSpy, 1)

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
	require.Equal(t, 1, rootSpy.beginTxCalls)
	assertMembershipExists(t, db, created.ID)
	assertMembershipBudget(t, db, semester.ID, 110)

	// A later composed operation fails after create has inserted the row and
	// credited the budget. The caller owns rollback of both effects.
	rollbackCreateUser := createMembershipTransactionUser(t, db)
	rollbackCreateTx, err := rootSpy.BeginTx()
	require.NoError(t, err)
	rollbackCreateSpy := &membershipPostgresLifecycleSpy{Store: rollbackCreateTx}
	rolledBackCreate, err := svc.CreateMembershipInTx(rollbackCreateSpy, semester.ID, &models.CreateMembershipRequest{
		UserID: rollbackCreateUser.ID,
		Paid:   true,
	}, models.MembershipSourceAdmin)
	require.NoError(t, err)
	insideSemester, err = rollbackCreateSpy.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(120), insideSemester.CurrentBudget, 0.001)
	duplicate := &models.Membership{UserID: rollbackCreateUser.ID, SemesterID: semester.ID}
	require.Error(t, rollbackCreateSpy.Memberships().Create(duplicate), "a later composed write should fail on the user/semester unique constraint")
	assertMembershipTxLifecycleUntouched(t, rootSpy, rollbackCreateSpy, 2)
	require.NoError(t, rollbackCreateSpy.Rollback())
	require.Equal(t, 1, rollbackCreateSpy.rollbackCalls)
	assertMembershipExternallyAbsent(t, db, rolledBackCreate.ID)
	assertMembershipBudget(t, db, semester.ID, 110)

	// Caller-owned update commits its membership transition and budget delta.
	updateCommitMembership := createMembershipTransactionMembership(t, db, semester.ID)
	updateCommitTx, err := rootSpy.BeginTx()
	require.NoError(t, err)
	updateCommitSpy := &membershipPostgresLifecycleSpy{Store: updateCommitTx}
	paid := true
	updated, err := svc.UpdateMembershipInTx(updateCommitSpy, updateCommitMembership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.True(t, updated.Paid)
	assertMembershipTxLifecycleUntouched(t, rootSpy, updateCommitSpy, 3)
	assertMembershipBudget(t, db, semester.ID, 110)
	require.NoError(t, updateCommitSpy.Commit())
	require.Zero(t, updateCommitSpy.beginTxCalls)
	require.Zero(t, updateCommitSpy.rollbackCalls)
	require.Equal(t, 1, updateCommitSpy.commitCalls)
	stored, err := postgres.NewStore(db).Memberships().FindByID(updateCommitMembership.ID)
	require.NoError(t, err)
	require.True(t, stored.Paid)
	assertMembershipBudget(t, db, semester.ID, 120)

	// The caller can also roll back a successful update and its budget delta.
	updateRollbackMembership := createMembershipTransactionMembership(t, db, semester.ID)
	updateRollbackTx, err := rootSpy.BeginTx()
	require.NoError(t, err)
	updateRollbackSpy := &membershipPostgresLifecycleSpy{Store: updateRollbackTx}
	updated, err = svc.UpdateMembershipInTx(updateRollbackSpy, updateRollbackMembership.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &paid})
	require.NoError(t, err)
	require.True(t, updated.Paid)
	insideSemester, err = updateRollbackSpy.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(130), insideSemester.CurrentBudget, 0.001)
	assertMembershipTxLifecycleUntouched(t, rootSpy, updateRollbackSpy, 4)
	require.NoError(t, updateRollbackSpy.Rollback())
	stored, err = postgres.NewStore(db).Memberships().FindByID(updateRollbackMembership.ID)
	require.NoError(t, err)
	require.False(t, stored.Paid)
	assertMembershipBudget(t, db, semester.ID, 120)

	// Standalone wrappers remain root-store entry points that commit on success.
	standaloneUser := createMembershipTransactionUser(t, db)
	standaloneCreated, err := svc.CreateMembership(semester.ID, &models.CreateMembershipRequest{UserID: standaloneUser.ID}, models.MembershipSourceAdmin)
	require.NoError(t, err)
	assertMembershipExists(t, db, standaloneCreated.ID)
	standalonePaid := true
	standaloneUpdated, err := svc.UpdateMembership(standaloneCreated.ID, semester.ID, &models.UpdateMembershipRequest{Paid: &standalonePaid})
	require.NoError(t, err)
	require.True(t, standaloneUpdated.Paid)
	stored, err = postgres.NewStore(db).Memberships().FindByID(standaloneCreated.ID)
	require.NoError(t, err)
	require.True(t, stored.Paid)
	assertMembershipBudget(t, db, semester.ID, 130)
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

func assertMembershipTxLifecycleUntouched(t *testing.T, root, tx *membershipPostgresLifecycleSpy, expectedRootBegins int) {
	t.Helper()
	require.Equal(t, expectedRootBegins, root.beginTxCalls, "injected methods must not begin a nested transaction")
	require.Zero(t, root.commitCalls, "injected methods must not commit the root store")
	require.Zero(t, root.rollbackCalls, "injected methods must not roll back the root store")
	require.Zero(t, tx.beginTxCalls, "injected methods must not begin a nested transaction")
	require.Zero(t, tx.commitCalls, "injected methods must not commit the caller transaction")
	require.Zero(t, tx.rollbackCalls, "injected methods must not roll back the caller transaction")
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
