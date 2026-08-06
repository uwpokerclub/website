package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/stretchr/testify/require"
)

func TestNewStore_Logins(t *testing.T) {
	t.Parallel()

	s := NewStore()

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, s.Logins().Create(login))

	found, err := s.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash1", found.Password)
}

func TestStore_BeginTx_Logins_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, tx.Logins().Create(login))

	// The parent store must not see the login until Commit is called.
	_, err = s.Logins().FindByUsername("alice")
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash1", found.Password)
}

func TestStore_BeginTx_Logins_DiscardedWithoutCommit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, tx.Logins().Create(login))
	require.NoError(t, tx.Rollback())

	_, err = s.Logins().FindByUsername("alice")
	require.ErrorIs(t, err, store.ErrNotFound)
}
