package controller_test

import (
	"api/internal/authorization"
	"api/internal/models"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEventClock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	semester, err := testutils.CreateTestSemester(db, "Fall 2025")
	require.NoError(t, err)
	structure, err := testutils.CreateTestStructure(db, "Standard")
	require.NoError(t, err)
	event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
	require.NoError(t, err)

	clockURL := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

	t.Run("GET requires authentication and at least executive", func(t *testing.T) {
		testutils.TestInvalidAuthForEndpoint(
			t, container, apiServer, "GET", clockURL, []string{"bot"},
		)
	})

	t.Run("control actions require at least tournament director", func(t *testing.T) {
		testutils.TestInvalidAuthForEndpoint(
			t, container, apiServer, "POST", clockURL+"/pause", []string{"bot", "executive"},
		)
	})

	t.Run("GET lazily creates a paused clock on first read", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "exec", authorization.ROLE_EXECUTIVE.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("GET", url, nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var body models.ClockState
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 0, body.LevelIndex)
		require.NotNil(t, body.PausedAt, "a freshly created clock starts paused")
		require.EqualValues(t, 1, body.Version)
	})

	t.Run("pause is a no-op when already paused", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", url+"/pause", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		var first models.ClockState
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &first))
		require.EqualValues(t, 1, first.Version, "the clock is created paused, so this is a lazy create only")

		req2, err := testutils.MakeJSONRequest("POST", url+"/pause", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req2, sessionID)
		w2 := httptest.NewRecorder()
		apiServer.ServeHTTP(w2, req2)
		require.Equal(t, http.StatusOK, w2.Code)
		var second models.ClockState
		require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &second))
		require.EqualValues(t, first.Version, second.Version, "pausing an already-paused clock must not bump version")
	})

	t.Run("resume unpauses and bumps version", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", url+"/resume", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		var body models.ClockState
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Nil(t, body.PausedAt)
		require.EqualValues(t, 2, body.Version, "lazy create (v1) plus the resume action itself (v2)")
	})

	t.Run("adjust validates deltaSeconds bounds", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", url+"/adjust", map[string]any{"deltaSeconds": 3601})
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		testutils.AssertBadRequestResponse(t, w, "")
	})

	t.Run("level rejects an out of range index", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", url+"/level", map[string]any{"index": 999})
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		testutils.AssertBadRequestResponse(t, w, "")
	})

	t.Run("level jumps to the requested index with a fresh duration", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", url+"/level", map[string]any{"index": 2})
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		var body models.ClockState
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 2, body.LevelIndex)
	})

	t.Run("control actions on an ended event return 409", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		require.NoError(t, db.Model(event).Update("state", models.EventStateEnded).Error)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", url+"/pause", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("a structure with no blinds has no clock", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure := &models.Structure{Name: "Empty"}
		require.NoError(t, db.Create(structure).Error)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "exec", authorization.ROLE_EXECUTIVE.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("GET", url, nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("concurrent adjusts serialize and none are lost", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Weekly")
		require.NoError(t, err)
		url := fmt.Sprintf("/api/v2/semesters/%s/events/%d/clock", semester.ID, event.ID)

		sessionID, err := testutils.CreateTestSession(db, "td", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
		require.NoError(t, err)

		// Establish the clock (version 1) before firing concurrent actions.
		getReq, err := testutils.MakeJSONRequest("GET", url, nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(getReq, sessionID)
		getW := httptest.NewRecorder()
		apiServer.ServeHTTP(getW, getReq)
		require.Equal(t, http.StatusOK, getW.Code)
		var initial models.ClockState
		require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &initial))

		const n = 5
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				req, err := testutils.MakeJSONRequest("POST", url+"/adjust", map[string]any{"deltaSeconds": 60})
				require.NoError(t, err)
				testutils.SetAuthCookie(req, sessionID)
				w := httptest.NewRecorder()
				apiServer.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
			}()
		}
		wg.Wait()

		finalReq, err := testutils.MakeJSONRequest("GET", url, nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(finalReq, sessionID)
		finalW := httptest.NewRecorder()
		apiServer.ServeHTTP(finalW, finalReq)
		require.Equal(t, http.StatusOK, finalW.Code)
		var final models.ClockState
		require.NoError(t, json.Unmarshal(finalW.Body.Bytes(), &final))

		require.EqualValues(t, initial.Version+n, final.Version, "every concurrent adjust must be reflected, none lost")
		want := initial.LevelEndsAt.Add(n * 60 * time.Second).Truncate(time.Microsecond)
		require.True(
			t,
			want.Equal(final.LevelEndsAt.Truncate(time.Microsecond)),
			"want %s, got %s", want, final.LevelEndsAt,
		)
	})
}
