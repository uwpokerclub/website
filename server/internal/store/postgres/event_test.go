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

func TestEventRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	repo := postgres.NewEventRepository(db)

	event := &models.Event{
		Name:             "Test Event",
		Format:           "No Limit Hold'em",
		SemesterID:       testutils.TEST_SEMESTERS[0].ID,
		StartDate:        time.Date(2024, 10, 1, 18, 0, 0, 0, time.UTC),
		StructureID:      testutils.TEST_STRUCTURES[0].ID,
		PointsMultiplier: 1.0,
	}

	require.NoError(t, repo.Create(event))
	require.NotZero(t, event.ID)
}

func TestEventRepository_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	repo := postgres.NewEventRepository(db)

	created, err := testutils.CreateTestEvent(
		db,
		testutils.TEST_SEMESTERS[0].ID,
		testutils.TEST_STRUCTURES[0].ID,
		"Findable Event",
	)
	require.NoError(t, err)

	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, "Findable Event", found.Name)
	require.NotNil(t, found.Semester)
	require.NotNil(t, found.Structure)
}

func TestEventRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	repo := postgres.NewEventRepository(container.GetDB())

	_, err = repo.FindByID(99999)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventRepository_FindBySemesterAndID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	repo := postgres.NewEventRepository(db)

	created, err := testutils.CreateTestEvent(
		db,
		testutils.TEST_SEMESTERS[0].ID,
		testutils.TEST_STRUCTURES[0].ID,
		"Scoped Event",
	)
	require.NoError(t, err)

	found, err := repo.FindBySemesterAndID(testutils.TEST_SEMESTERS[0].ID, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, "Scoped Event", found.Name)
	require.NotNil(t, found.Semester)
	require.NotNil(t, found.Structure)

	_, err = repo.FindBySemesterAndID(testutils.TEST_SEMESTERS[1].ID, created.ID)
	require.ErrorIs(t, err, store.ErrNotFound)

	_, err = repo.FindBySemesterAndID(testutils.TEST_SEMESTERS[0].ID, 99999)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	repo := postgres.NewEventRepository(db)

	semesterID := testutils.TEST_SEMESTERS[0].ID
	_, err = testutils.CreateTestEvent(db, semesterID, testutils.TEST_STRUCTURES[0].ID, "Alpha Tournament")
	require.NoError(t, err)
	_, err = testutils.CreateTestEvent(db, semesterID, testutils.TEST_STRUCTURES[0].ID, "Beta Tournament")
	require.NoError(t, err)
	_, err = testutils.CreateTestEvent(
		db,
		testutils.TEST_SEMESTERS[1].ID,
		testutils.TEST_STRUCTURES[0].ID,
		"Other Semester Event",
	)
	require.NoError(t, err)

	events, total, err := repo.List(&models.ListEventsFilter{SemesterID: &semesterID})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, events, 2)

	events, total, err = repo.List(&models.ListEventsFilter{SemesterID: &semesterID, Search: "Alpha"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, events, 1)
	require.Equal(t, "Alpha Tournament", events[0].Name)
}

func TestEventRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedSemesters(db))
	require.NoError(t, testutils.SeedStructures(db))

	repo := postgres.NewEventRepository(db)

	created, err := testutils.CreateTestEvent(
		db,
		testutils.TEST_SEMESTERS[0].ID,
		testutils.TEST_STRUCTURES[0].ID,
		"Original Name",
	)
	require.NoError(t, err)

	err = repo.Update(created, map[string]any{"name": "Updated Name"})
	require.NoError(t, err)
	require.Equal(t, "Updated Name", created.Name)

	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", found.Name)
}

func TestEventRepository_List_AllSemesters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	structure := &models.Structure{Name: "Standard"}
	require.NoError(t, db.Create(structure).Error)
	semesterA := &models.Semester{Name: "Fall 2026"}
	require.NoError(t, db.Create(semesterA).Error)
	semesterB := &models.Semester{Name: "Winter 2026"}
	require.NoError(t, db.Create(semesterB).Error)

	repo := postgres.NewEventRepository(db)
	require.NoError(t, repo.Create(&models.Event{Name: "A", Format: "NL", SemesterID: semesterA.ID, StructureID: structure.ID, StartDate: time.Now().UTC()}))
	require.NoError(t, repo.Create(&models.Event{Name: "B", Format: "NL", SemesterID: semesterB.ID, StructureID: structure.ID, StartDate: time.Now().UTC()}))

	results, total, err := repo.List(&models.ListEventsFilter{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, int64(2))
	require.GreaterOrEqual(t, len(results), 2)

	results, total, err = repo.List(&models.ListEventsFilter{SemesterID: &semesterA.ID})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, results, 1)
	require.Equal(t, "A", results[0].Name)
}

func TestEventsEntries_CascadeOnEventDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	require.NoError(t, testutils.SeedAll(db))

	// TEST_EVENTS id 2 has two seeded participants (TEST_PARTICIPANTS[2] and [3]).
	var before int64
	require.NoError(t, db.Model(&models.Participant{}).Where("event_id = ?", 2).Count(&before).Error)
	require.EqualValues(t, 2, before)

	require.NoError(t, db.Delete(&models.Event{}, "id = ?", 2).Error)

	var after int64
	require.NoError(t, db.Model(&models.Participant{}).Where("event_id = ?", 2).Count(&after).Error)
	require.EqualValues(t, 0, after)

	// Participants belonging to other events must survive.
	var untouched int64
	require.NoError(t, db.Model(&models.Participant{}).Where("event_id = ?", 1).Count(&untouched).Error)
	require.EqualValues(t, 2, untouched)
}
