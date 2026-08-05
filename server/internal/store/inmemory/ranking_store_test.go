package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewStore_Rankings(t *testing.T) {
	t.Parallel()

	s := NewStore()

	ranking := &models.Ranking{MembershipID: uuid.New(), Points: 10}
	require.NoError(t, s.Rankings().Create(ranking))

	found, err := s.Rankings().FindByMembershipID(ranking.MembershipID)
	require.NoError(t, err)
	require.Equal(t, ranking.ID, found.ID)
}

func TestStore_BeginTx_Rankings_Commit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	ranking := &models.Ranking{MembershipID: uuid.New(), Points: 10}
	require.NoError(t, tx.Rankings().Create(ranking))

	// The parent store must not see the ranking until Commit is called.
	_, err = s.Rankings().FindByMembershipID(ranking.MembershipID)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, tx.Commit())

	found, err := s.Rankings().FindByMembershipID(ranking.MembershipID)
	require.NoError(t, err)
	require.Equal(t, ranking.ID, found.ID)
}

func TestStore_BeginTx_Rankings_DiscardedWithoutCommit(t *testing.T) {
	t.Parallel()

	s := NewStore()

	tx, err := s.BeginTx()
	require.NoError(t, err)

	ranking := &models.Ranking{MembershipID: uuid.New(), Points: 10}
	require.NoError(t, tx.Rankings().Create(ranking))
	require.NoError(t, tx.Rollback())

	_, err = s.Rankings().FindByMembershipID(ranking.MembershipID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
