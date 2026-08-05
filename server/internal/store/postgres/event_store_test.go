package postgres_test

import (
	"context"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/stretchr/testify/require"
)

func TestNewStore_Events(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	s := postgres.NewStore(db)

	event := &models.Event{
		Name:             "Store Wiring Event",
		Format:           "No Limit Hold'em",
		SemesterID:       testutils.TEST_SEMESTERS[0].ID,
		StartDate:        time.Date(2024, 10, 1, 18, 0, 0, 0, time.UTC),
		StructureID:      testutils.TEST_STRUCTURES[0].ID,
		PointsMultiplier: 1.0,
	}
	require.NoError(t, s.Events().Create(event))

	found, err := s.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, event.ID, found.ID)
}

func TestStore_BeginTx_Events_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	s := postgres.NewStore(db)

	tx, err := s.BeginTx()
	require.NoError(t, err)

	event := &models.Event{
		Name:             "Rolled Back Event",
		Format:           "No Limit Hold'em",
		SemesterID:       testutils.TEST_SEMESTERS[0].ID,
		StartDate:        time.Date(2024, 10, 1, 18, 0, 0, 0, time.UTC),
		StructureID:      testutils.TEST_STRUCTURES[0].ID,
		PointsMultiplier: 1.0,
	}
	require.NoError(t, tx.Events().Create(event))
	require.NoError(t, tx.Rollback())

	_, err = s.Events().FindByID(event.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
