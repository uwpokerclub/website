package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/stretchr/testify/require"
)

func TestLoginRepository_Create(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	login := &models.Login{Username: "alice", Password: "hash1", Role: "executive"}
	require.NoError(t, repo.Create(login))
}

func TestLoginRepository_Create_DuplicateUsername(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	err := repo.Create(&models.Login{Username: "alice", Password: "hash2", Role: "president"})
	require.Error(t, err)
}

func TestLoginRepository_FindByUsername(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	found, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash1", found.Password)
	require.Equal(t, "executive", found.Role)

	_, err = repo.FindByUsername("nobody")
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Update(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	require.NoError(t, repo.Update("alice", map[string]any{"role": "president"}))

	found, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "president", found.Role)
	require.Equal(t, "hash1", found.Password)
}

func TestLoginRepository_Update_PartialFields(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	require.NoError(t, repo.Update("alice", map[string]any{"password": "hash2"}))

	found, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "hash2", found.Password)
	require.Equal(t, "executive", found.Role)
}

func TestLoginRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	err := repo.Update("nobody", map[string]any{"role": "president"})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Delete(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))
	require.NoError(t, repo.Delete("alice"))

	_, err := repo.FindByUsername("alice")
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	err := repo.Delete("nobody")
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestLoginRepository_Clone_Isolation(t *testing.T) {
	t.Parallel()

	repo := newLoginRepository()

	require.NoError(t, repo.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"}))

	clone := repo.clone()

	// Mutating the clone must not affect the original.
	require.NoError(t, clone.Update("alice", map[string]any{"role": "president"}))

	original, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "executive", original.Role)

	cloned, err := clone.FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "president", cloned.Role)

	// Creating a new login on the original must not appear in the clone.
	require.NoError(t, repo.Create(&models.Login{Username: "bob", Password: "hash2", Role: "secretary"}))

	_, err = clone.FindByUsername("bob")
	require.ErrorIs(t, err, store.ErrNotFound)
	require.Len(t, clone.logins, 1)
	require.Len(t, repo.logins, 2)
}
