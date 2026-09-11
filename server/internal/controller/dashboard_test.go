package controller_test

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// createDashboardTestEvent creates an event with fully custom state and start date, since
// testutils.CreateTestEvent always creates a started event dated at time.Now().
func createDashboardTestEvent(
	db *gorm.DB,
	semesterID uuid.UUID,
	structureID int32,
	name string,
	state uint8,
	startDate time.Time,
	rebuys uint8,
) (*models.Event, error) {
	event := models.Event{
		Name:             name,
		Format:           "No Limit Hold'em",
		SemesterID:       semesterID,
		StartDate:        startDate,
		State:            state,
		StructureID:      structureID,
		Rebuys:           rebuys,
		PointsMultiplier: 1.0,
	}

	if err := db.Create(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func TestDashboardSpotlight(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	semester, err := testutils.CreateTestSemester(db, "Fall 2025")
	require.NoError(t, err)

	t.Run("requires authentication and at least executive", func(t *testing.T) {
		testutils.TestInvalidAuthForEndpoint(
			t, container, apiServer, "GET",
			fmt.Sprintf("/api/v2/semesters/%s/dashboard/spotlight", semester.ID),
			[]string{"bot"},
		)
	})

	getSpotlight := func(t *testing.T, semesterID string) *httptest.ResponseRecorder {
		sessionID, err := testutils.CreateTestSession(db, "exec-"+uuid.NewString(), "executive")
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest(
			"GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/spotlight", semesterID), nil,
		)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		return w
	}

	t.Run("returns the live event with its entry count and rebuys", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		event, err := createDashboardTestEvent(
			db, semester.ID, structure.ID, "Weekly",
			models.EventStateStarted, time.Now().Add(-time.Hour), 5,
		)
		require.NoError(t, err)

		user1, err := testutils.CreateTestUser(db, 11111111, "A", "One", "a@uwaterloo.ca", models.FacultyMath, "a1")
		require.NoError(t, err)
		user2, err := testutils.CreateTestUser(db, 22222222, "B", "Two", "b@uwaterloo.ca", models.FacultyMath, "b2")
		require.NoError(t, err)
		membership1, err := testutils.CreateTestMembership(db, user1.ID, semester.ID)
		require.NoError(t, err)
		membership2, err := testutils.CreateTestMembership(db, user2.ID, semester.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership1.ID, event.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership2.ID, event.ID)
		require.NoError(t, err)

		w := getSpotlight(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body store.SpotlightEvent
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, event.ID, body.ID)
		require.Equal(t, "Weekly", body.Name)
		require.EqualValues(t, 2, body.Entries)
		require.EqualValues(t, 5, body.Rebuys)
		require.EqualValues(t, models.EventStateStarted, body.State)
	})

	t.Run("a live event takes priority over a future event", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		liveEvent, err := createDashboardTestEvent(
			db, semester.ID, structure.ID, "Live", models.EventStateStarted, time.Now().Add(-time.Hour), 0,
		)
		require.NoError(t, err)
		_, err = createDashboardTestEvent(
			db, semester.ID, structure.ID, "Future", models.EventStateStarted, time.Now().Add(time.Hour), 0,
		)
		require.NoError(t, err)

		w := getSpotlight(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body store.SpotlightEvent
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, liveEvent.ID, body.ID)
	})

	t.Run("returns the soonest event when only future events exist", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		soonest, err := createDashboardTestEvent(
			db, semester.ID, structure.ID, "Soonest", models.EventStateStarted, time.Now().Add(time.Hour), 0,
		)
		require.NoError(t, err)
		_, err = createDashboardTestEvent(
			db, semester.ID, structure.ID, "Later", models.EventStateStarted, time.Now().Add(48*time.Hour), 0,
		)
		require.NoError(t, err)

		w := getSpotlight(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body store.SpotlightEvent
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, soonest.ID, body.ID)
	})

	t.Run("returns null when the semester has no events", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)

		w := getSpotlight(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "null", w.Body.String())
	})

	t.Run("returns null when every event has ended", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		_, err = createDashboardTestEvent(
			db, semester.ID, structure.ID, "Past", models.EventStateEnded, time.Now().Add(-time.Hour), 0,
		)
		require.NoError(t, err)

		w := getSpotlight(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "null", w.Body.String())
	})

	t.Run("entry count excludes entries belonging to other events", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		liveEvent, err := createDashboardTestEvent(
			db, semester.ID, structure.ID, "Live", models.EventStateStarted, time.Now().Add(-time.Hour), 0,
		)
		require.NoError(t, err)
		otherEvent, err := createDashboardTestEvent(
			db, semester.ID, structure.ID, "Other", models.EventStateEnded, time.Now().Add(-48*time.Hour), 0,
		)
		require.NoError(t, err)

		user1, err := testutils.CreateTestUser(db, 11111111, "A", "One", "a@uwaterloo.ca", models.FacultyMath, "a1")
		require.NoError(t, err)
		user2, err := testutils.CreateTestUser(db, 22222222, "B", "Two", "b@uwaterloo.ca", models.FacultyMath, "b2")
		require.NoError(t, err)
		membership1, err := testutils.CreateTestMembership(db, user1.ID, semester.ID)
		require.NoError(t, err)
		membership2, err := testutils.CreateTestMembership(db, user2.ID, semester.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership1.ID, liveEvent.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership2.ID, otherEvent.ID)
		require.NoError(t, err)

		w := getSpotlight(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body store.SpotlightEvent
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, liveEvent.ID, body.ID)
		require.EqualValues(t, 1, body.Entries)
	})

	t.Run("returns 404 for an unknown semester id", func(t *testing.T) {
		w := getSpotlight(t, "00000000-0000-0000-0000-000000000000")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for a malformed semester id", func(t *testing.T) {
		w := getSpotlight(t, "not-a-uuid")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
