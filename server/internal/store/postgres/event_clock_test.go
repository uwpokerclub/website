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
	"gorm.io/gorm"
)

// setupEventClockFixtures spins up a container with a semester, structure,
// and event already created, satisfying event_clocks' FK to events.
func setupEventClockFixtures(t *testing.T) (db *gorm.DB, eventID int32) {
	t.Helper()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { container.Close(ctx) })

	db = container.GetDB()

	semester, err := testutils.CreateTestSemester(db, "Event Clock Test Semester")
	require.NoError(t, err)

	structure, err := testutils.CreateTestStructure(db, "Event Clock Test Structure")
	require.NoError(t, err)

	event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Event Clock Test Event")
	require.NoError(t, err)

	return db, event.ID
}

func TestEventClockRepository_FindByEventID_NotFound(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	_, err := repo.FindByEventID(eventID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_Create_FindByEventID(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := &models.EventClock{
		EventID:     eventID,
		LevelIndex:  0,
		LevelEndsAt: now.Add(15 * time.Minute),
		PausedAt:    &now,
		Version:     1,
		UpdatedAt:   now,
	}

	require.NoError(t, repo.Create(clock))

	found, err := repo.FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, clock.EventID, found.EventID)
	require.Equal(t, clock.LevelIndex, found.LevelIndex)
	require.WithinDuration(t, clock.LevelEndsAt, found.LevelEndsAt, time.Millisecond)
	require.NotNil(t, found.PausedAt)
	require.WithinDuration(t, *clock.PausedAt, *found.PausedAt, time.Millisecond)
	require.Equal(t, clock.Version, found.Version)
}

func TestEventClockRepository_Create_AlreadyExists(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.EventClock{EventID: eventID, LevelEndsAt: now, Version: 1, UpdatedAt: now}))

	err := repo.Create(&models.EventClock{EventID: eventID, LevelEndsAt: now, Version: 1, UpdatedAt: now})
	require.ErrorIs(t, err, store.ErrAlreadyExists)
}

func TestEventClockRepository_Update(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := &models.EventClock{EventID: eventID, LevelIndex: 0, LevelEndsAt: now.Add(15 * time.Minute), PausedAt: &now, Version: 1, UpdatedAt: now}
	require.NoError(t, repo.Create(clock))

	later := now.Add(time.Hour)
	clock.LevelIndex = 2
	clock.LevelEndsAt = later.Add(15 * time.Minute)
	clock.PausedAt = nil
	clock.Version = 2
	clock.UpdatedAt = later

	require.NoError(t, repo.Update(clock))

	found, err := repo.FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, int32(2), found.LevelIndex)
	require.Nil(t, found.PausedAt, "Update must clear a nil PausedAt, not skip it as a zero value")
	require.Equal(t, int64(2), found.Version)
}

func TestEventClockRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	err := repo.Update(&models.EventClock{EventID: eventID})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_DeleteByEventID(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.EventClock{EventID: eventID, LevelEndsAt: now, Version: 1, UpdatedAt: now}))

	require.NoError(t, repo.DeleteByEventID(eventID))

	_, err := repo.FindByEventID(eventID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_DeleteByEventID_NotFound(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	err := repo.DeleteByEventID(eventID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventClockRepository_FindByEventIDForUpdate_BlocksConcurrentTx(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	repo := postgres.NewEventClockRepository(db)

	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.EventClock{EventID: eventID, LevelEndsAt: now, Version: 1, UpdatedAt: now}))

	tx1 := db.Begin()
	t.Cleanup(func() { tx1.Rollback() })
	tx1Repo := postgres.NewEventClockRepository(tx1)

	_, err := tx1Repo.FindByEventIDForUpdate(eventID)
	require.NoError(t, err)

	// A second transaction's locking read on the same row must block until
	// tx1 releases the lock.
	secondStarted := make(chan struct{})
	secondFinished := make(chan struct{})
	go func() {
		tx2 := db.Begin()
		defer tx2.Rollback()
		tx2Repo := postgres.NewEventClockRepository(tx2)

		close(secondStarted)
		_, err := tx2Repo.FindByEventIDForUpdate(eventID)
		require.NoError(t, err)
		close(secondFinished)
	}()

	<-secondStarted
	select {
	case <-secondFinished:
		t.Fatal("second transaction's locking read must block while the first holds the lock")
	case <-time.After(200 * time.Millisecond):
	}

	require.NoError(t, tx1.Commit().Error)

	select {
	case <-secondFinished:
	case <-time.After(5 * time.Second):
		t.Fatal("second transaction's locking read never unblocked after the first committed")
	}
}
