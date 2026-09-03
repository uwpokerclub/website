package postgres_test

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/postgres"

	"github.com/stretchr/testify/require"
)

func TestNewStore_EventClocks(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	s := postgres.NewStore(db)

	now := time.Now().UTC()
	clock := &models.EventClock{EventID: eventID, LevelEndsAt: now, Version: 1, UpdatedAt: now}
	require.NoError(t, s.EventClocks().Create(clock))

	found, err := s.EventClocks().FindByEventID(eventID)
	require.NoError(t, err)
	require.Equal(t, clock.EventID, found.EventID)
}

func TestStore_BeginTx_EventClocks_Rollback(t *testing.T) {
	t.Parallel()

	db, eventID := setupEventClockFixtures(t)
	s := postgres.NewStore(db)

	tx, err := s.BeginTx()
	require.NoError(t, err)

	now := time.Now().UTC()
	require.NoError(t, tx.EventClocks().Create(&models.EventClock{EventID: eventID, LevelEndsAt: now, Version: 1, UpdatedAt: now}))
	require.NoError(t, tx.Rollback())

	_, err = s.EventClocks().FindByEventID(eventID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
