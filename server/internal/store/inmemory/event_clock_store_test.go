package inmemory

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"

	"github.com/stretchr/testify/require"
)

func TestNewStore_EventClocks(t *testing.T) {
	t.Parallel()

	s := NewStore()

	now := time.Now().UTC()
	clock := &models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now}
	require.NoError(t, s.EventClocks().Create(clock))

	found, err := s.EventClocks().FindByEventID(1)
	require.NoError(t, err)
	require.Equal(t, clock.EventID, found.EventID)
}

func TestStore_BeginTx_EventClocks_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	now := time.Now().UTC()
	clock := &models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now}
	require.NoError(t, tx.EventClocks().Create(clock))

	// The parent store must not see the clock until Commit is called.
	_, err = s.EventClocks().FindByEventID(1)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.EventClocks().FindByEventID(1)
	require.NoError(t, err)
	require.Equal(t, clock.EventID, found.EventID)
}

func TestStore_BeginTx_EventClocks_Rollback(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	now := time.Now().UTC()
	require.NoError(t, tx.EventClocks().Create(&models.EventClock{EventID: 1, LevelEndsAt: now, Version: 1, UpdatedAt: now}))
	require.NoError(t, tx.Rollback())

	_, err = s.EventClocks().FindByEventID(1)
	require.ErrorIs(t, err, store.ErrNotFound)
}
