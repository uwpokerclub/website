package postgres_test

import (
	"context"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}).Error)

	repo := postgres.NewSessionRepository(db)

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))
	require.NotEqual(t, uuid.Nil, session.ID)
}

func TestSessionRepository_Create_UnknownUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewSessionRepository(container.GetDB())

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "nobody", Role: "executive"}
	require.Error(t, repo.Create(session))
}

func TestSessionRepository_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}).Error)

	repo := postgres.NewSessionRepository(db)

	now := time.Now().UTC()
	expiresAt := now.Add(time.Hour)
	session := &models.Session{StartedAt: now, ExpiresAt: expiresAt, Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))

	found, err := repo.FindByID(session.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", found.Username)
	require.Equal(t, "executive", found.Role)
	require.WithinDuration(t, expiresAt, found.ExpiresAt, time.Second)
}

func TestSessionRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewSessionRepository(container.GetDB())

	_, err = repo.FindByID(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestSessionRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}).Error)

	repo := postgres.NewSessionRepository(db)

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))

	require.NoError(t, repo.Delete(session.ID))

	_, err = repo.FindByID(session.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestSessionRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewSessionRepository(container.GetDB())

	err = repo.Delete(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestSessionRepository_DeleteByUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}).Error)
	require.NoError(t, db.Create(&models.Login{Username: "bob", Password: "hash2", Role: "executive"}).Error)
	repo := postgres.NewSessionRepository(db)
	now := time.Now().UTC()
	aliceOne := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	aliceTwo := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "president"}
	bob := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "bob", Role: "executive"}
	require.NoError(t, repo.Create(aliceOne))
	require.NoError(t, repo.Create(aliceTwo))
	require.NoError(t, repo.Create(bob))

	require.NoError(t, repo.DeleteByUsername("alice"))
	_, err = repo.FindByID(aliceOne.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
	_, err = repo.FindByID(aliceTwo.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
	_, err = repo.FindByID(bob.ID)
	require.NoError(t, err)
}
