package inmemory

import (
	"testing"

	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewStore_Events(t *testing.T) {
	t.Parallel()

	s := NewStore()
	semesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "Store Wiring Event")
	require.NoError(t, s.Events().Create(event))

	found, err := s.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, event.ID, found.ID)
}

func TestStore_BeginTx_Events_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()
	semesterID := uuid.New()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	event := newTestEvent(semesterID, 1, "Committed Event")
	require.NoError(t, tx.Events().Create(event))

	// The parent store must not see the event until Commit is called.
	_, err = s.Events().FindByID(event.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.Events().FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, event.ID, found.ID)
}

func TestStore_BeginTx_Events_DiscardedWithoutCommit(t *testing.T) {
	t.Parallel()

	s := NewStore()
	semesterID := uuid.New()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	event := newTestEvent(semesterID, 1, "Uncommitted Event")
	require.NoError(t, tx.Events().Create(event))
	require.NoError(t, tx.Rollback())

	_, err = s.Events().FindByID(event.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
