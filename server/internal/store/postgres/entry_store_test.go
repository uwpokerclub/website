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

func TestNewStore_Entries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Store Wiring Event")
	require.NoError(t, err)
	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	s := postgres.NewStore(db)

	participant := &models.Participant{
		MembershipID: &membership.ID,
		EventID:      event.ID,
	}
	require.NoError(t, s.Entries().Create(participant))

	found, err := s.Entries().FindByID(participant.ID)
	require.NoError(t, err)
	require.Equal(t, participant.ID, found.ID)
}

func TestStore_BeginTx_Entries_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Rollback Event")
	require.NoError(t, err)
	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	s := postgres.NewStore(db)

	tx, err := s.BeginTx()
	require.NoError(t, err)

	participant := &models.Participant{
		MembershipID: &membership.ID,
		EventID:      event.ID,
	}
	require.NoError(t, tx.Entries().Create(participant))
	require.NoError(t, tx.Rollback())

	_, err = s.Entries().FindByID(participant.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
