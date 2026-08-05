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
