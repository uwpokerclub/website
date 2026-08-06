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

func TestSemesterRepository_IncrementBudget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	repo := postgres.NewSemesterRepository(db)

	semester := &models.Semester{Name: "Fall 2026", StartingBudget: 100, CurrentBudget: 100}
	require.NoError(t, repo.Create(semester))

	require.NoError(t, repo.IncrementBudget(semester.ID, 25))
	found, err := repo.FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(125), found.CurrentBudget, 0.001)

	require.NoError(t, repo.IncrementBudget(semester.ID, -50))
	found, err = repo.FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(75), found.CurrentBudget, 0.001)
}

func TestSemesterRepository_IncrementBudget_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewSemesterRepository(container.GetDB())

	err = repo.IncrementBudget(uuid.New(), 10)
	require.ErrorIs(t, err, store.ErrNotFound)
}
