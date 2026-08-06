package inmemory

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository_Create(t *testing.T) {
	t.Parallel()

	repo := newSessionRepository()

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))
	require.NotEqual(t, uuid.Nil, session.ID)
}

func TestSessionRepository_Create_DuplicateID(t *testing.T) {
	t.Parallel()

	repo := newSessionRepository()

	id := uuid.New()
	now := time.Now().UTC()
	require.NoError(t, repo.Create(&models.Session{ID: id, StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}))

	dup := &models.Session{ID: id, StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "bob", Role: "executive"}
	require.Error(t, repo.Create(dup))
}

func TestSessionRepository_FindByID(t *testing.T) {
	t.Parallel()

	repo := newSessionRepository()

	now := time.Now().UTC()
	expiresAt := now.Add(time.Hour)
	session := &models.Session{StartedAt: now, ExpiresAt: expiresAt, Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))

	found, err := repo.FindByID(session.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", found.Username)
	require.Equal(t, "executive", found.Role)
	require.Equal(t, expiresAt, found.ExpiresAt)

	_, err = repo.FindByID(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestSessionRepository_Delete(t *testing.T) {
	t.Parallel()

	repo := newSessionRepository()

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))
	require.NoError(t, repo.Delete(session.ID))

	_, err := repo.FindByID(session.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestSessionRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	repo := newSessionRepository()

	err := repo.Delete(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestSessionRepository_Clone_Isolation(t *testing.T) {
	t.Parallel()

	repo := newSessionRepository()

	now := time.Now().UTC()
	session := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "alice", Role: "executive"}
	require.NoError(t, repo.Create(session))

	clone := repo.clone()

	// Deleting from the clone must not affect the original.
	require.NoError(t, clone.Delete(session.ID))

	_, err := repo.FindByID(session.ID)
	require.NoError(t, err)

	_, err = clone.FindByID(session.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	// Creating a new session on the original must not appear in the clone.
	other := &models.Session{StartedAt: now, ExpiresAt: now.Add(time.Hour), Username: "bob", Role: "executive"}
	require.NoError(t, repo.Create(other))

	_, err = clone.FindByID(other.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
	require.Len(t, clone.sessions, 0)
	require.Len(t, repo.sessions, 2)
}
