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

func TestEntryRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	participant := &models.Participant{
		MembershipID: &membership.ID,
		EventID:      event.ID,
	}

	require.NoError(t, repo.Create(participant))
	require.NotZero(t, participant.ID)
}

func TestEntryRepository_Create_DuplicateMembershipEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	_, err = testutils.CreateTestParticipant(db, membership.ID, event.ID)
	require.NoError(t, err)

	dup := &models.Participant{
		MembershipID: &membership.ID,
		EventID:      event.ID,
	}
	require.Error(t, repo.Create(dup))
}

func TestEntryRepository_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	created, err := testutils.CreateTestParticipant(db, membership.ID, event.ID)
	require.NoError(t, err)

	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, event.ID, found.EventID)
	require.Nil(t, found.Membership)
}

func TestEntryRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewEntryRepository(container.GetDB())

	_, err = repo.FindByID(99999)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_FindByMembershipAndEventID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	eventA, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Event A")
	require.NoError(t, err)
	eventB, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Event B")
	require.NoError(t, err)

	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	created, err := testutils.CreateTestParticipant(db, membership.ID, eventA.ID)
	require.NoError(t, err)

	found, err := repo.FindByMembershipAndEventID(membership.ID, eventA.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)

	_, err = repo.FindByMembershipAndEventID(membership.ID, eventB.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	_, err = repo.FindByMembershipAndEventID(uuid.New(), eventA.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)
	otherEvent, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Other Event")
	require.NoError(t, err)

	membershipA, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)
	membershipB, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[1].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)
	membershipC, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[2].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	// Not yet signed out -- must sort first.
	stillIn, err := testutils.CreateTestParticipant(db, membershipA.ID, event.ID)
	require.NoError(t, err)

	// Signed out earlier.
	signedOutEarlier, err := testutils.CreateTestParticipant(db, membershipB.ID, event.ID)
	require.NoError(t, err)
	earlier := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	require.NoError(t, db.Model(signedOutEarlier).Update("signed_out_at", earlier).Error)

	// Signed out later -- must sort between "still in" and "signed out earlier".
	signedOutLater, err := testutils.CreateTestParticipant(db, membershipC.ID, event.ID)
	require.NoError(t, err)
	later := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Model(signedOutLater).Update("signed_out_at", later).Error)

	// Belongs to a different event -- must be excluded.
	_, err = testutils.CreateTestParticipant(db, membershipA.ID, otherEvent.ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	participants, total, err := repo.List(&models.ListParticipantsFilter{EventID: event.ID})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, participants, 3)

	require.Equal(t, stillIn.ID, participants[0].ID)
	require.Equal(t, signedOutLater.ID, participants[1].ID)
	require.Equal(t, signedOutEarlier.ID, participants[2].ID)

	require.NotNil(t, participants[0].Membership)
	require.NotNil(t, participants[0].Membership.User)
}

func TestEntryRepository_Update_SignOut(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)
	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	created, err := testutils.CreateTestParticipant(db, membership.ID, event.ID)
	require.NoError(t, err)
	require.Nil(t, created.SignedOutAt)

	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, repo.Update(created, map[string]any{"signed_out_at": &now}))
	require.NotNil(t, created.SignedOutAt)
	require.WithinDuration(t, now, *created.SignedOutAt, time.Second)

	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, found.SignedOutAt)
}

func TestEntryRepository_Update_SignIn(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)
	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	created, err := testutils.CreateTestParticipant(db, membership.ID, event.ID)
	require.NoError(t, err)
	now := time.Now().UTC()
	require.NoError(t, db.Model(created).Update("signed_out_at", now).Error)

	require.NoError(t, repo.Update(created, map[string]any{"signed_out_at": nil}))
	require.Nil(t, created.SignedOutAt)

	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	require.Nil(t, found.SignedOutAt)
}

func TestEntryRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))
	require.NoError(t, testutils.SeedUsers(db))

	event, err := testutils.CreateTestEvent(db, testutils.TEST_SEMESTERS[0].ID, testutils.TEST_STRUCTURES[0].ID, "Test Event")
	require.NoError(t, err)
	membership, err := testutils.CreateTestMembership(db, testutils.TEST_USERS[0].ID, testutils.TEST_SEMESTERS[0].ID)
	require.NoError(t, err)

	repo := postgres.NewEntryRepository(db)

	_, err = testutils.CreateTestParticipant(db, membership.ID, event.ID)
	require.NoError(t, err)

	require.NoError(t, repo.Delete(membership.ID, event.ID))

	_, err = repo.FindByMembershipAndEventID(membership.ID, event.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewEntryRepository(container.GetDB())

	err = repo.Delete(uuid.New(), 99999)
	require.ErrorIs(t, err, store.ErrNotFound)
}
