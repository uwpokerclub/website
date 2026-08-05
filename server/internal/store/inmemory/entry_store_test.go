package inmemory

import (
	"testing"

	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewStore_Entries(t *testing.T) {
	t.Parallel()

	s := NewStore()

	participant := newTestParticipant(uuid.New(), 1)
	require.NoError(t, s.Entries().Create(participant))

	found, err := s.Entries().FindByID(participant.ID)
	require.NoError(t, err)
	require.Equal(t, participant.ID, found.ID)
}

func TestStore_BeginTx_Entries_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	participant := newTestParticipant(uuid.New(), 1)
	require.NoError(t, tx.Entries().Create(participant))

	// The parent store must not see the entry until Commit is called.
	_, err = s.Entries().FindByID(participant.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.Entries().FindByID(participant.ID)
	require.NoError(t, err)
	require.Equal(t, participant.ID, found.ID)
}

func TestStore_BeginTx_Entries_DiscardedWithoutCommit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	participant := newTestParticipant(uuid.New(), 1)
	require.NoError(t, tx.Entries().Create(participant))
	require.NoError(t, tx.Rollback())

	_, err = s.Entries().FindByID(participant.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
