package controller_test

import (
	"api/internal/controller"
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

// createDashboardTestSemester creates a semester with a custom start date, since
// testutils.CreateTestSemester always uses a fixed Fall 2024 date.
func createDashboardTestSemester(db *gorm.DB, name string, startDate time.Time) (*models.Semester, error) {
	semester := models.Semester{
		Name:                  name,
		Meta:                  "Test semester",
		StartDate:             startDate,
		EndDate:               startDate.AddDate(0, 3, 0),
		StartingBudget:        100.0,
		CurrentBudget:         100.0,
		MembershipFee:         10,
		MembershipDiscountFee: 5,
		RebuyFee:              2,
	}

	if err := db.Create(&semester).Error; err != nil {
		return nil, err
	}

	return &semester, nil
}

// createDashboardTestMembership creates a membership with custom paid/discounted/
// executive flags, since testutils.CreateTestMembership always creates a plain paid
// membership.
func createDashboardTestMembership(
	db *gorm.DB, userID uint64, semesterID uuid.UUID, paid, discounted, executive bool,
) (*models.Membership, error) {
	membership := models.Membership{
		UserID:     userID,
		SemesterID: semesterID,
		Paid:       paid,
		Discounted: discounted,
		Executive:  executive,
	}

	if err := db.Create(&membership).Error; err != nil {
		return nil, err
	}

	return &membership, nil
}

func TestDashboardMembershipStats(t *testing.T) {
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
			fmt.Sprintf("/api/v2/semesters/%s/dashboard/memberships", semester.ID),
			[]string{"bot"},
		)
	})

	getMembershipStats := func(t *testing.T, semesterID string) *httptest.ResponseRecorder {
		sessionID, err := testutils.CreateTestSession(db, "exec-"+uuid.NewString(), "executive")
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest(
			"GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/memberships", semesterID), nil,
		)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		return w
	}

	nextUserID := func() uint64 {
		return uint64(time.Now().UnixNano())
	}

	t.Run("buckets are an exclusive partition that sums to the total", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		paidUser, err := testutils.CreateTestUser(db, nextUserID(), "Paid", "User", "paid@uwaterloo.ca", models.FacultyMath, "p1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, paidUser.ID, semester.ID, true, false, false)
		require.NoError(t, err)

		discountedUser, err := testutils.CreateTestUser(db, nextUserID(), "Discounted", "User", "disc@uwaterloo.ca", models.FacultyMath, "d1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, discountedUser.ID, semester.ID, true, true, false)
		require.NoError(t, err)

		unpaidUser, err := testutils.CreateTestUser(db, nextUserID(), "Unpaid", "User", "unpaid@uwaterloo.ca", models.FacultyMath, "u1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, unpaidUser.ID, semester.ID, false, false, false)
		require.NoError(t, err)

		execUser, err := testutils.CreateTestUser(db, nextUserID(), "Exec", "User", "exec@uwaterloo.ca", models.FacultyMath, "e1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, execUser.ID, semester.ID, false, false, true)
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 4, body.Current.Total)
		require.EqualValues(t, 1, body.Current.Paid)
		require.EqualValues(t, 1, body.Current.Discounted)
		require.EqualValues(t, 1, body.Current.Unpaid)
		require.EqualValues(t, 1, body.Current.Executive)
		require.Equal(t, body.Current.Total, body.Current.Paid+body.Current.Unpaid+body.Current.Discounted+body.Current.Executive)
	})

	t.Run("an unpaid executive lands only in executive", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		execUser, err := testutils.CreateTestUser(db, nextUserID(), "Exec", "User", "exec@uwaterloo.ca", models.FacultyMath, "e1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, execUser.ID, semester.ID, false, false, true)
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.Total)
		require.EqualValues(t, 1, body.Current.Executive)
		require.EqualValues(t, 0, body.Current.Unpaid)
		require.EqualValues(t, 0, body.Current.Paid)
		require.EqualValues(t, 0, body.Current.Discounted)
	})

	t.Run("an unpaid discounted membership lands in unpaid", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		user, err := testutils.CreateTestUser(db, nextUserID(), "Free", "Trial", "trial@uwaterloo.ca", models.FacultyMath, "t1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, semester.ID, false, true, false)
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.Unpaid)
		require.EqualValues(t, 0, body.Current.Discounted)
	})

	t.Run("a member whose first-ever membership is this term counts as new", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		user, err := testutils.CreateTestUser(db, nextUserID(), "New", "Member", "new@uwaterloo.ca", models.FacultyMath, "n1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.New)
		require.EqualValues(t, 0, body.Current.Returning)
	})

	t.Run("a member with a membership in an earlier semester counts as returning", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		earlier, err := createDashboardTestSemester(db, "Winter 2025", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		current, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		user, err := testutils.CreateTestUser(db, nextUserID(), "Returning", "Member", "returning@uwaterloo.ca", models.FacultyMath, "r1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, earlier.ID, true, false, false)
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, current.ID, true, false, false)
		require.NoError(t, err)

		w := getMembershipStats(t, current.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.Total)
		require.EqualValues(t, 0, body.Current.New)
		require.EqualValues(t, 1, body.Current.Returning)
	})

	t.Run("each semester's new/returning split is measured from its own start date", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		comparisonSemester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		user, err := testutils.CreateTestUser(db, nextUserID(), "CrossTerm", "Member", "crossterm@uwaterloo.ca", models.FacultyMath, "c1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, comparisonSemester.ID, true, false, false)
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, currentSemester.ID, true, false, false)
		require.NoError(t, err)

		w := getMembershipStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		require.Equal(t, comparisonSemester.ID, body.Comparison.Semester.ID)

		require.EqualValues(t, 0, body.Current.New)
		require.EqualValues(t, 1, body.Current.Returning)

		require.EqualValues(t, 1, body.Comparison.Stats.New)
		require.EqualValues(t, 0, body.Comparison.Stats.Returning)
	})

	t.Run("a member created in the gap between terms with no earlier membership still counts as new", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		// A user created well before the semester's start date, with no membership until now -
		// proves the split is keyed off membership history, not users.created_at.
		user, err := testutils.CreateTestUser(db, nextUserID(), "GapSignup", "User", "gap@uwaterloo.ca", models.FacultyMath, "g1")
		require.NoError(t, err)
		require.NoError(t, db.Model(user).Update("created_at", time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)).Error)
		_, err = createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.New)
		require.EqualValues(t, 0, body.Current.Returning)
	})

	t.Run("returns zeroes, not an error, for a semester with no memberships", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 0, body.Current.Total)
		require.EqualValues(t, 0, body.Current.New)
		require.EqualValues(t, 0, body.Current.Returning)
	})

	t.Run("returns comparison null when there is no comparable prior term", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		w := getMembershipStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Nil(t, body.Comparison)
	})

	t.Run("returns the comparison semester's own independently correct stats", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		comparisonSemester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		execUser, err := testutils.CreateTestUser(db, nextUserID(), "Exec", "User", "exec@uwaterloo.ca", models.FacultyMath, "e1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, execUser.ID, comparisonSemester.ID, false, false, true)
		require.NoError(t, err)
		paidUser, err := testutils.CreateTestUser(db, nextUserID(), "Paid", "User", "paid@uwaterloo.ca", models.FacultyMath, "p1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, paidUser.ID, comparisonSemester.ID, true, false, false)
		require.NoError(t, err)

		w := getMembershipStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		require.Equal(t, "Fall 2025", body.Comparison.Semester.Name)
		require.EqualValues(t, 2, body.Comparison.Stats.Total)
		require.EqualValues(t, 1, body.Comparison.Stats.Executive)
		require.EqualValues(t, 1, body.Comparison.Stats.Paid)
		require.EqualValues(t, 0, body.Current.Total)
	})

	t.Run("returns 404 for an unknown semester id", func(t *testing.T) {
		w := getMembershipStats(t, "00000000-0000-0000-0000-000000000000")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for a malformed semester id", func(t *testing.T) {
		w := getMembershipStats(t, "not-a-uuid")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
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
