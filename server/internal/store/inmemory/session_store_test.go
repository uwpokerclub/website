package inmemory

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"

	"github.com/stretchr/testify/require"
)

func TestNewStore_Sessions(t *testing.T) {
	t.Parallel()

	s := NewStore()

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, s.Sessions().Create(session))

	found, err := s.Sessions().FindByID(session.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", found.Username)
}

func TestStore_BeginTx_Sessions_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, tx.Sessions().Create(session))

	// The parent store must not see the session until Commit is called.
	_, err = s.Sessions().FindByID(session.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.Sessions().FindByID(session.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", found.Username)
}

func TestStore_BeginTx_Sessions_DiscardedWithoutCommit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, tx.Sessions().Create(session))
	require.NoError(t, tx.Rollback())

	_, err = s.Sessions().FindByID(session.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
