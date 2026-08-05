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

func TestLoginRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, repo.Create(login))
}

func TestLoginRepository_Create_DuplicateUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	dup := &models.Login{Username: "alice", Password: "hash2", Role: "president"}
	require.Error(t, repo.Create(dup))
}

func TestLoginRepository_FindByUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	found, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash1", found.Password)
	require.Equal(t, "executive", found.Role)
}

func TestLoginRepository_FindByUsername_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	_, err = repo.FindByUsername("nobody")
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	require.NoError(t, repo.Update("alice", map[string]any{"role": "president"}))

	found, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "president", found.Role)
	require.Equal(t, "hash1", found.Password)
}

func TestLoginRepository_Update_PartialFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	// Updating only password must leave role untouched.
	require.NoError(t, repo.Update("alice", map[string]any{"password": "hash2"}))

	found, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash2", found.Password)
	require.Equal(t, "executive", found.Role)
}

func TestLoginRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	err = repo.Update("nobody", map[string]any{"role": "president"})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))
	require.NoError(t, repo.Delete("alice"))

	_, err = repo.FindByUsername("alice")
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewLoginRepository(container.GetDB())

	err = repo.Delete("nobody")
	require.ErrorIs(t, err, store.ErrNotFound)
}
