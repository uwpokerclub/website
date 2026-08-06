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

func TestLoginRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, db.Create(&models.User{ID: 20260010, FirstName: "Alan", LastName: "Turing", Email: "alan@example.com", Faculty: "Math", QuestID: "aturing"}).Error)

	repo := postgres.NewLoginRepository(db)
	require.NoError(t, repo.Create(&models.Login{Username: "aturing", Password: "hash", Role: "executive"}))
	require.NoError(t, repo.Create(&models.Login{Username: "bot", Password: "hash", Role: "bot"}))

	results, total, err := repo.List(&models.Pagination{}, "")
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, results, 2)

	var withMember *models.LoginWithMember
	for i := range results {
		if results[i].Username == "aturing" {
			withMember = &results[i]
		}
	}
	require.NotNil(t, withMember)
	require.NotNil(t, withMember.LinkedMember)
	require.Equal(t, "Alan", withMember.LinkedMember.FirstName)

	results, total, err = repo.List(&models.Pagination{}, "turing")
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	require.Equal(t, "aturing", results[0].Username)
}

func TestLoginRepository_FindByUsernameWithMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, db.Create(&models.User{ID: 20260011, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com", Faculty: "Math", QuestID: "alovelace"}).Error)

	repo := postgres.NewLoginRepository(db)
	require.NoError(t, repo.Create(&models.Login{Username: "alovelace", Password: "hash", Role: "executive"}))
	require.NoError(t, repo.Create(&models.Login{Username: "unlinked", Password: "hash", Role: "bot"}))

	found, err := repo.FindByUsernameWithMember("alovelace")
	require.NoError(t, err)
	require.NotNil(t, found.LinkedMember)
	require.Equal(t, "Ada", found.LinkedMember.FirstName)

	found, err = repo.FindByUsernameWithMember("unlinked")
	require.NoError(t, err)
	require.Nil(t, found.LinkedMember)

	_, err = repo.FindByUsernameWithMember("nobody")
	require.ErrorIs(t, err, store.ErrNotFound)
}
