package postgres_test

import (
	"context"
	"testing"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMembershipRepository_CreateAndFindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	repo := postgres.NewMembershipRepository(db)

	user, err := testutils.CreateTestUser(db, 20000001, "Jane", "Doe", "jane@example.com", "Math", "jdoe")
	require.NoError(t, err)

	semester, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	membership := &models.Membership{
		UserID:     user.ID,
		SemesterID: semester.ID,
		Paid:       true,
		Discounted: false,
	}

	require.NoError(t, repo.Create(membership))
	require.NotEqual(t, uuid.Nil, membership.ID)

	found, err := repo.FindByID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, membership.ID, found.ID)
	require.Equal(t, user.ID, found.UserID)
	require.Equal(t, semester.ID, found.SemesterID)
	require.True(t, found.Paid)
	require.False(t, found.Discounted)
}

func TestMembershipRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewMembershipRepository(container.GetDB())

	_, err = repo.FindByID(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_FindByIDAndSemesterID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	repo := postgres.NewMembershipRepository(db)

	user, err := testutils.CreateTestUser(db, 20000002, "John", "Smith", "john@example.com", "Science", "jsmith")
	require.NoError(t, err)

	semesterA, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	semesterB, err := testutils.CreateTestSemester(db, "Winter 2027")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, user.ID, semesterA.ID)
	require.NoError(t, err)

	found, err := repo.FindByIDAndSemesterID(membership.ID, semesterA.ID)
	require.NoError(t, err)
	require.Equal(t, membership.ID, found.ID)

	_, err = repo.FindByIDAndSemesterID(membership.ID, semesterB.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_FindByID_Preloads(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	user, err := testutils.CreateTestUser(db, 20260001, "Ada", "Lovelace", "ada@example.com", "Math", "alovelace")
	require.NoError(t, err)
	semester, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	repo := postgres.NewMembershipRepository(db)
	membership := &models.Membership{UserID: user.ID, SemesterID: semester.ID, Paid: true}
	require.NoError(t, repo.Create(membership))

	found, err := repo.FindByID(membership.ID)
	require.NoError(t, err)
	require.NotNil(t, found.User)
	require.Equal(t, "Ada", found.User.FirstName)
	require.NotNil(t, found.Semester)
	require.Equal(t, "Fall 2026", found.Semester.Name)
}

func TestMembershipRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	repo := postgres.NewMembershipRepository(db)

	semesterA, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	semesterB, err := testutils.CreateTestSemester(db, "Winter 2027")
	require.NoError(t, err)

	user1, err := testutils.CreateTestUser(db, 20000010, "Alice", "A", "alice@example.com", "Math", "aa")
	require.NoError(t, err)
	user2, err := testutils.CreateTestUser(db, 20000011, "Bob", "B", "bob@example.com", "Math", "bb")
	require.NoError(t, err)
	user3, err := testutils.CreateTestUser(db, 20000012, "Cara", "C", "cara@example.com", "Math", "cc")
	require.NoError(t, err)

	m1 := &models.Membership{UserID: user1.ID, SemesterID: semesterA.ID, Paid: true, Discounted: false}
	m2 := &models.Membership{UserID: user2.ID, SemesterID: semesterA.ID, Paid: false, Discounted: false}
	m3 := &models.Membership{UserID: user3.ID, SemesterID: semesterB.ID, Paid: true, Discounted: false}
	require.NoError(t, repo.Create(m1))
	require.NoError(t, repo.Create(m2))
	require.NoError(t, repo.Create(m3))

	structure, err := testutils.CreateTestStructure(db, "Standard")
	require.NoError(t, err)
	event, err := testutils.CreateTestEvent(db, semesterA.ID, structure.ID, "Weekly")
	require.NoError(t, err)
	_, err = testutils.CreateTestParticipant(db, m1.ID, event.ID)
	require.NoError(t, err)

	// Filter by semester only: expect m1 and m2, ordered by first/last name ascending, with
	// m1's attendance reflecting the one event it participated in.
	results, total, err := repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA.ID})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, results, 2)
	require.Equal(t, user1.ID, results[0].UserID)
	require.Equal(t, 1, results[0].Attendance)
	require.Equal(t, user2.ID, results[1].UserID)
	require.Equal(t, 0, results[1].Attendance)

	// Filter by semester and paid status: expect only m1.
	paidTrue := true
	results, total, err = repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA.ID, Paid: &paidTrue})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	require.Equal(t, m1.ID, results[0].ID)

	// Pagination: limit 1, offset 1 within semester A returns the second member (Bob).
	limit := 1
	offset := 1
	results, total, err = repo.List(&models.ListMembershipsFilter{
		SemesterID: &semesterA.ID,
		Pagination: models.Pagination{Limit: &limit, Offset: &offset},
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, results, 1)
	require.Equal(t, user2.ID, results[0].UserID)

	// Name filter matching only Bob.
	name := "bob"
	results, total, err = repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA.ID, Name: &name})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	require.Equal(t, user2.ID, results[0].UserID)

	// Name filter matching nobody.
	noMatch := "nobody"
	results, total, err = repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA.ID, Name: &noMatch})
	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, results)
}

func TestMembershipRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	repo := postgres.NewMembershipRepository(db)

	user, err := testutils.CreateTestUser(db, 20000020, "Dana", "D", "dana@example.com", "Math", "dd")
	require.NoError(t, err)

	semester, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, user.ID, semester.ID)
	require.NoError(t, err)

	membership.Paid = false
	membership.Discounted = false
	require.NoError(t, repo.Update(membership))

	found, err := repo.FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, found.Paid)
	require.False(t, found.Discounted)
}

func TestMembershipRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewMembershipRepository(container.GetDB())

	err = repo.Update(&models.Membership{ID: uuid.New()})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	repo := postgres.NewMembershipRepository(db)

	user, err := testutils.CreateTestUser(db, 20000030, "Eli", "E", "eli@example.com", "Math", "ee")
	require.NoError(t, err)

	semesterA, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	semesterB, err := testutils.CreateTestSemester(db, "Winter 2027")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, user.ID, semesterA.ID)
	require.NoError(t, err)

	// Deleting with the wrong semester ID should not find the row.
	err = repo.Delete(membership.ID, semesterB.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, repo.Delete(membership.ID, semesterA.ID))

	_, err = repo.FindByID(membership.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	err = repo.Delete(membership.ID, semesterA.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
