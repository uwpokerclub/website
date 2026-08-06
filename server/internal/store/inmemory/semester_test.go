package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSemesterRepository_IncrementBudget(t *testing.T) {
	t.Parallel()

	repo := newSemesterRepository()

	semester := &models.Semester{StartingBudget: 100, CurrentBudget: 100}
	require.NoError(t, repo.Create(semester))

	require.NoError(t, repo.IncrementBudget(semester.ID, 25))
	found, err := repo.FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(125), found.CurrentBudget, 0.001)
}

func TestSemesterRepository_IncrementBudget_NotFound(t *testing.T) {
	t.Parallel()

	repo := newSemesterRepository()

	err := repo.IncrementBudget(uuid.New(), 10)
	require.ErrorIs(t, err, store.ErrNotFound)
}
