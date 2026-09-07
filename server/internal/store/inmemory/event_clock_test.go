package inmemory

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"

	"github.com/stretchr/testify/require"
)

func TestEventClockRepository_FindByEventID_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	_, err := repo.FindByEventID(1)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_Create_FindByEventID(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	now := time.Now().UTC()
	clock := &models.EventClock{
		EventID:     1,
		LevelIndex:  0,
		LevelEndsAt: now.Add(15 * time.Minute),
		PausedAt:    &now,
		Version:     1,
		UpdatedAt:   now,
	}

	require.NoError(t, repo.Create(clock))

	found, err := repo.FindByEventID(1)
	require.NoError(t, err)
	require.Equal(t, clock.EventID, found.EventID)
	require.Equal(t, clock.LevelIndex, found.LevelIndex)
	require.WithinDuration(t, clock.LevelEndsAt, found.LevelEndsAt, 0)
	require.NotNil(t, found.PausedAt)
	require.WithinDuration(t, *clock.PausedAt, *found.PausedAt, 0)
	require.Equal(t, clock.Version, found.Version)
}

func TestEventClockRepository_Create_AlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now}))

	err := repo.Create(&models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now})
	require.ErrorIs(t, err, store.ErrAlreadyExists)
}

func TestEventClockRepository_Update(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	now := time.Now().UTC()
	clock := &models.EventClock{EventID: 1, LevelIndex: 0, LevelEndsAt: now.Add(15 * time.Minute), PausedAt: &now, Version: 1, UpdatedAt: now}
	require.NoError(t, repo.Create(clock))

	later := now.Add(time.Hour)
	clock.LevelIndex = 2
	clock.LevelEndsAt = later.Add(15 * time.Minute)
	clock.PausedAt = nil
	clock.Version = 2
	clock.UpdatedAt = later

	require.NoError(t, repo.Update(clock))

	found, err := repo.FindByEventID(1)
	require.NoError(t, err)
	require.Equal(t, int32(2), found.LevelIndex)
	require.Nil(t, found.PausedAt, "Update must clear a nil PausedAt, not skip it as a zero value")
	require.Equal(t, int64(2), found.Version)
}

func TestEventClockRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	err := repo.Update(&models.EventClock{EventID: 1})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_DeleteByEventID(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now}))

	require.NoError(t, repo.DeleteByEventID(1))

	_, err := repo.FindByEventID(1)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_DeleteByEventID_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	err := repo.DeleteByEventID(1)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_FindByEventIDForUpdate(t *testing.T) {
	t.Parallel()

	repo := NewEventClockRepository()

	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now}))

	found, err := repo.FindByEventIDForUpdate(1)
	require.NoError(t, err)
	require.Equal(t, int32(1), found.EventID)
}
