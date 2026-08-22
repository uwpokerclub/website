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

func TestRankingRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)

	ranking := &models.Ranking{
		MembershipID: membership.ID,
		Points:       10,
	}

	require.NoError(t, repo.Create(ranking))
	require.NotZero(t, ranking.ID)
}

func TestRankingRepository_Create_DuplicateMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)

	require.NoError(t, repo.Create(&models.Ranking{MembershipID: membership.ID, Points: 5}))

	dup := &models.Ranking{MembershipID: membership.ID, Points: 1}
	require.Error(t, repo.Create(dup))
}

func TestRankingRepository_FindByMembershipID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)

	created := &models.Ranking{MembershipID: membership.ID, Points: 15}
	require.NoError(t, repo.Create(created))

	found, err := repo.FindByMembershipID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.EqualValues(t, 15, found.Points)
}

func TestRankingRepository_FindByMembershipID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewRankingRepository(container.GetDB())

	_, err = repo.FindByMembershipID(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestRankingRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)

	created := &models.Ranking{MembershipID: membership.ID, Points: 5}
	require.NoError(t, repo.Create(created))

	created.Points = 20
	require.NoError(t, repo.Update(created))

	found, err := repo.FindByMembershipID(membership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 20, found.Points)
}

func TestRankingRepository_Update_ToZero(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)

	created := &models.Ranking{MembershipID: membership.ID, Points: 5}
	require.NoError(t, repo.Create(created))

	// Regression guard: Points=0 is GORM's zero value and must not be silently
	// dropped by Updates -- Update must use Select("points") to force it through.
	created.Points = 0
	require.NoError(t, repo.Update(created))

	found, err := repo.FindByMembershipID(membership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 0, found.Points)
}

func TestRankingRepository_BatchIncrementPoints_CreatesForNewMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{membership.ID: 12}))

	found, err := repo.FindByMembershipID(membership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 12, found.Points)
}

func TestRankingRepository_BatchIncrementPoints_IncrementsExisting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: membership.ID, Points: 8}))

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{membership.ID: 5}))

	found, err := repo.FindByMembershipID(membership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 13, found.Points)
}

func TestRankingRepository_BatchIncrementPoints_Mixed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedUsers(db))

	existingMembership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)
	newMembership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[1].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewRankingRepository(db)
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: existingMembership.ID, Points: 3}))

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{
		existingMembership.ID: 7,
		newMembership.ID:      4,
	}))

	foundExisting, err := repo.FindByMembershipID(existingMembership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 10, foundExisting.Points)

	foundNew, err := repo.FindByMembershipID(newMembership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 4, foundNew.Points)
}

func TestRankingRepository_BatchIncrementPoints_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewRankingRepository(container.GetDB())

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{}))
}

func TestRankingRepository_FindBySemesterAndMembershipID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	user := &models.User{ID: 20260030, FirstName: "Marie", LastName: "Curie", Email: "mc@example.com", Faculty: "Science"}
	require.NoError(t, db.Create(user).Error)
	semester := &models.Semester{Name: "Fall 2026"}
	require.NoError(t, db.Create(semester).Error)
	membership := &models.Membership{UserID: user.ID, SemesterID: semester.ID}
	require.NoError(t, db.Create(membership).Error)

	repo := postgres.NewRankingRepository(db)
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: membership.ID, Points: 42}))

	found, err := repo.FindBySemesterAndMembershipID(semester.ID, membership.ID)
	require.NoError(t, err)
	require.EqualValues(t, 42, found.Points)
	require.EqualValues(t, 1, found.Position)

	_, err = repo.FindBySemesterAndMembershipID(semester.ID, uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestRankingRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	user := &models.User{ID: 20260031, FirstName: "Rosalind", LastName: "Franklin", Email: "rf@example.com", Faculty: "Science"}
	require.NoError(t, db.Create(user).Error)
	semester := &models.Semester{Name: "Winter 2026"}
	require.NoError(t, db.Create(semester).Error)
	membership := &models.Membership{UserID: user.ID, SemesterID: semester.ID}
	require.NoError(t, db.Create(membership).Error)

	repo := postgres.NewRankingRepository(db)
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: membership.ID, Points: 10}))

	results, total, err := repo.List(&models.ListRankingsFilter{SemesterID: semester.ID})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	require.Equal(t, "Rosalind", results[0].FirstName)

	results, total, err = repo.List(&models.ListRankingsFilter{SemesterID: semester.ID, Search: "franklin"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)

	results, total, err = repo.List(&models.ListRankingsFilter{SemesterID: semester.ID, Search: "nobody"})
	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, results)
}
