package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewStore_Memberships(t *testing.T) {
	t.Parallel()

	s := NewStore()

	membership := &models.Membership{UserID: 1, SemesterID: uuid.New(), Paid: true}
	require.NoError(t, s.Memberships().Create(membership))

	found, err := s.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, membership.ID, found.ID)
}

func TestStore_BeginTx_Memberships_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	membership := &models.Membership{UserID: 1, SemesterID: uuid.New(), Paid: true}
	require.NoError(t, tx.Memberships().Create(membership))

	// The parent store must not see the membership until Commit is called.
	_, err = s.Memberships().FindByID(membership.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, membership.ID, found.ID)
}

func TestStore_BeginTx_Memberships_DiscardedWithoutCommit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	membership := &models.Membership{UserID: 1, SemesterID: uuid.New(), Paid: true}
	require.NoError(t, tx.Memberships().Create(membership))
	require.NoError(t, tx.Rollback())

	_, err = s.Memberships().FindByID(membership.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
