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
)

func TestNewStore_Sessions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	s := postgres.NewStore(container.GetDB())

	require.NoError(t, s.Logins().Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, s.Sessions().Create(session))

	found, err := s.Sessions().FindByID(session.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", found.Username)
}

func TestStore_BeginTx_Sessions_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	s := postgres.NewStore(container.GetDB())

	require.NoError(t, s.Logins().Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	tx, err := s.BeginTx()
	require.NoError(t, err)

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, tx.Sessions().Create(session))
	require.NoError(t, tx.Rollback())

	_, err = s.Sessions().FindByID(session.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
