package postgres_test

import (
	"context"
	"testing"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/stretchr/testify/require"
)

func TestNewStore_Rankings(t *testing.T) {
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

	s := postgres.NewStore(db)

	ranking := &models.Ranking{MembershipID: membership.ID, Points: 10}
	require.NoError(t, s.Rankings().Create(ranking))

	found, err := s.Rankings().FindByMembershipID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, ranking.ID, found.ID)
}

func TestStore_BeginTx_Rankings_Rollback(t *testing.T) {
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

	s := postgres.NewStore(db)

	tx, err := s.BeginTx()
	require.NoError(t, err)

	ranking := &models.Ranking{MembershipID: membership.ID, Points: 10}
	require.NoError(t, tx.Rankings().Create(ranking))
	require.NoError(t, tx.Rollback())

	_, err = s.Rankings().FindByMembershipID(membership.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
