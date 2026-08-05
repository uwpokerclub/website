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

func TestNewStore_Memberships(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	s := postgres.NewStore(db)

	user, err := testutils.CreateTestUser(db, 20000040, "Fay", "F", "fay@example.com", "Math", "ff")
	require.NoError(t, err)

	semester, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	membership := &models.Membership{UserID: user.ID, SemesterID: semester.ID, Paid: true}
	require.NoError(t, s.Memberships().Create(membership))

	found, err := s.Memberships().FindByID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, membership.ID, found.ID)
}

func TestStore_BeginTx_Memberships_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	s := postgres.NewStore(db)

	user, err := testutils.CreateTestUser(db, 20000041, "Gus", "G", "gus@example.com", "Math", "gg")
	require.NoError(t, err)

	semester, err := testutils.CreateTestSemester(db, "Fall 2026")
	require.NoError(t, err)

	tx, err := s.BeginTx()
	require.NoError(t, err)

	membership := &models.Membership{UserID: user.ID, SemesterID: semester.ID, Paid: true}
	require.NoError(t, tx.Memberships().Create(membership))
	require.NoError(t, tx.Rollback())

	_, err = s.Memberships().FindByID(membership.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
