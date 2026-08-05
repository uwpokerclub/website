package inmemory

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestEvent(semesterID uuid.UUID, structureID int32, name string) *models.Event {
	return &models.Event{
		Name:             name,
		Format:           "No Limit Hold'em",
		SemesterID:       semesterID,
		StartDate:        time.Date(2024, 10, 1, 18, 0, 0, 0, time.UTC),
		StructureID:      structureID,
		PointsMultiplier: 1.0,
	}
}

func TestEventRepository_Create(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "Test Event")
	require.NoError(t, repo.Create(event))
	require.NotZero(t, event.ID)
}

func TestEventRepository_FindByID(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "Findable Event")
	require.NoError(t, repo.Create(event))

	found, err := repo.FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, "Findable Event", found.Name)

	_, err = repo.FindByID(99999)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventRepository_Create_DuplicateID(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "First")
	require.NoError(t, repo.Create(event))

	dup := newTestEvent(semesterID, 1, "Second")
	dup.ID = event.ID
	err := repo.Create(dup)
	require.Error(t, err)
}

func TestEventRepository_FindBySemesterAndID(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()
	otherSemesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "Scoped Event")
	require.NoError(t, repo.Create(event))

	found, err := repo.FindBySemesterAndID(semesterID, event.ID)
	require.NoError(t, err)
	require.Equal(t, event.ID, found.ID)

	_, err = repo.FindBySemesterAndID(otherSemesterID, event.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventRepository_List(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()

	require.NoError(t, repo.Create(newTestEvent(semesterID, 1, "Alpha Tournament")))
	require.NoError(t, repo.Create(newTestEvent(semesterID, 1, "Beta Tournament")))
	require.NoError(t, repo.Create(newTestEvent(uuid.New(), 1, "Other Semester Event")))

	events, total, err := repo.List(&models.ListEventsFilter{SemesterID: semesterID})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, events, 2)

	events, total, err = repo.List(&models.ListEventsFilter{SemesterID: semesterID, Search: "alpha"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "Alpha Tournament", events[0].Name)
}

func TestEventRepository_Update(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "Original Name")
	require.NoError(t, repo.Create(event))

	err := repo.Update(event, map[string]any{
		"name":              "Updated Name",
		"points_multiplier": float32(2.0),
	})
	require.NoError(t, err)
	require.Equal(t, "Updated Name", event.Name)
	require.Equal(t, float32(2.0), event.PointsMultiplier)

	found, err := repo.FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", found.Name)
}

func TestEventRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()

	err := repo.Update(&models.Event{ID: 99999}, map[string]any{"name": "X"})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEventRepository_Clone_Isolation(t *testing.T) {
	t.Parallel()

	repo := newEventRepository()
	semesterID := uuid.New()

	event := newTestEvent(semesterID, 1, "Original Name")
	require.NoError(t, repo.Create(event))

	clone := repo.clone()

	// Mutating the clone must not affect the original.
	require.NoError(t, clone.Update(&models.Event{ID: event.ID}, map[string]any{"name": "Cloned Name"}))

	original, err := repo.FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, "Original Name", original.Name)

	cloned, err := clone.FindByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, "Cloned Name", cloned.Name)

	// Creating a new event on the original must not appear in the clone.
	require.NoError(t, repo.Create(newTestEvent(semesterID, 1, "Second Event")))

	_, cloneTotal, err := clone.List(&models.ListEventsFilter{SemesterID: semesterID})
	require.NoError(t, err)
	require.EqualValues(t, 1, cloneTotal)

	_, originalTotal, err := repo.List(&models.ListEventsFilter{SemesterID: semesterID})
	require.NoError(t, err)
	require.EqualValues(t, 2, originalTotal)
}
