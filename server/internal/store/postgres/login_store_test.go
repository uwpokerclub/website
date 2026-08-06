package postgres_test

import (
	"context"
	"testing"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/stretchr/testify/require"
)

func TestNewStore_Logins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	s := postgres.NewStore(container.GetDB())

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, s.Logins().Create(login))

	found, err := s.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash1", found.Password)
}

func TestStore_BeginTx_Logins_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	s := postgres.NewStore(container.GetDB())

	tx, err := s.BeginTx()
	require.NoError(t, err)

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, tx.Logins().Create(login))
	require.NoError(t, tx.Rollback())

	_, err = s.Logins().FindByUsername("alice")
	require.ErrorIs(t, err, store.ErrNotFound)
}
