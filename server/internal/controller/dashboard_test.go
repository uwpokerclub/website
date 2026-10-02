package controller_test

import (
	"api/internal/controller"
	"api/internal/models"
	"api/internal/store"
	postgresstore "api/internal/store/postgres"
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

	t.Run("reports totalAsOf null when the comparison term has no dated memberships", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		comparisonSemester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		user, err := testutils.CreateTestUser(db, nextUserID(), "Undated", "User", "undated@uwaterloo.ca", models.FacultyMath, "u1")
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, comparisonSemester.ID, true, false, false)
		require.NoError(t, err)
		// GORM populates CreatedAt on insert, so a pre-migration row has to be made
		// explicitly: every membership predating #429 carries NULL here.
		require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", membership.ID).
			Update("created_at", nil).Error)

		w := getMembershipStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		// Null, never zero: the term has memberships, their creation dates are simply
		// unknown, and reporting 0 would read as "nobody had signed up by now".
		require.Nil(t, body.Comparison.TotalAsOf)
		require.EqualValues(t, 1, body.Comparison.Stats.Total)
	})

	t.Run("withholds totalAsOf when most of the comparison term is undated", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()

		now := time.Now().UTC()
		currentStart := now.AddDate(0, 0, -20)
		comparisonStart := currentStart.AddDate(-1, 0, 0)

		comparisonSemester, err := createDashboardTestSemester(db, "Prior", comparisonStart)
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Current", currentStart)
		require.NoError(t, err)

		// The shape of the term that straddled the migration: one dated row among many.
		for i := 0; i < 10; i++ {
			user, err := testutils.CreateTestUser(db, nextUserID(), "Member", fmt.Sprintf("%d", i),
				fmt.Sprintf("s%d@uwaterloo.ca", i), models.FacultyMath, fmt.Sprintf("s%d", i))
			require.NoError(t, err)
			membership, err := createDashboardTestMembership(db, user.ID, comparisonSemester.ID, true, false, false)
			require.NoError(t, err)

			value := any(nil)
			if i == 0 {
				value = comparisonStart.AddDate(0, 0, 1)
			}
			require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", membership.ID).
				Update("created_at", value).Error)
		}

		w := getMembershipStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		// A baseline of 1 would have made any current term look like a triumph.
		require.Nil(t, body.Comparison.TotalAsOf)
		require.EqualValues(t, 10, body.Comparison.Stats.Total)
	})

	t.Run("counts only comparison memberships created by the same elapsed point", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()

		now := time.Now().UTC()
		currentStart := now.AddDate(0, 0, -20)
		comparisonStart := currentStart.AddDate(-1, 0, 0)

		comparisonSemester, err := createDashboardTestSemester(db, "Prior", comparisonStart)
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Current", currentStart)
		require.NoError(t, err)

		for i, created := range []time.Time{
			comparisonStart.AddDate(0, 0, 2),  // inside the 20-day window
			comparisonStart.AddDate(0, 0, 5),  // inside
			comparisonStart.AddDate(0, 0, 60), // after it
		} {
			user, err := testutils.CreateTestUser(db, nextUserID(), "Member", fmt.Sprintf("%d", i),
				fmt.Sprintf("m%d@uwaterloo.ca", i), models.FacultyMath, fmt.Sprintf("m%d", i))
			require.NoError(t, err)
			membership, err := createDashboardTestMembership(db, user.ID, comparisonSemester.ID, true, false, false)
			require.NoError(t, err)
			require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", membership.ID).
				Update("created_at", created).Error)
		}

		w := getMembershipStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.MembershipStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		require.NotNil(t, body.Comparison.TotalAsOf)
		require.EqualValues(t, 2, *body.Comparison.TotalAsOf)
		// The final total still carries all three, so the card can show pace as well.
		require.EqualValues(t, 3, body.Comparison.Stats.Total)
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

func TestDashboardEngagementStats(t *testing.T) {
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
			fmt.Sprintf("/api/v2/semesters/%s/dashboard/engagement", semester.ID),
			[]string{"bot"},
		)
	})

	getEngagementStats := func(t *testing.T, semesterID string) *httptest.ResponseRecorder {
		sessionID, err := testutils.CreateTestSession(db, "exec-"+uuid.NewString(), "executive")
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest(
			"GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/engagement", semesterID), nil,
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

	// entrant creates a user, a membership in semesterID, and enters them into each of
	// the given events.
	entrant := func(t *testing.T, db *gorm.DB, semesterID uuid.UUID, tag string, eventIDs ...int32) *models.Membership {
		user, err := testutils.CreateTestUser(db, nextUserID(), tag, "Player", tag+"@uwaterloo.ca", models.FacultyMath, tag)
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, semesterID, true, false, false)
		require.NoError(t, err)
		for _, eventID := range eventIDs {
			_, err = testutils.CreateTestParticipant(db, membership.ID, eventID)
			require.NoError(t, err)
		}
		return membership
	}

	t.Run("a member with a membership but zero entries is excluded from every figure", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E1", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)

		entrant(t, db, semester.ID, "player", event.ID)

		// Membership with no entries - must not appear in any figure.
		bystander, err := testutils.CreateTestUser(db, nextUserID(), "No", "Entries", "noentries@uwaterloo.ca", models.FacultyMath, "ne1")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, bystander.ID, semester.ID, true, false, false)
		require.NoError(t, err)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.Players)
	})

	t.Run("median is correct for an odd population", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		e1, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E1", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)
		e2, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E2", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)
		e3, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E3", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)

		// Three players with 1, 2, and 3 events attended - median is 2.
		entrant(t, db, semester.ID, "p1", e1.ID)
		entrant(t, db, semester.ID, "p2", e1.ID, e2.ID)
		entrant(t, db, semester.ID, "p3", e1.ID, e2.ID, e3.ID)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 3, body.Current.Players)
		require.Equal(t, 2.0, body.Current.MedianEventsAttended)
	})

	t.Run("median is correct for an even population", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		e1, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E1", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)
		e2, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E2", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)

		// Two players with 1 and 2 events attended - median is 1.5.
		entrant(t, db, semester.ID, "p1", e1.ID)
		entrant(t, db, semester.ID, "p2", e1.ID, e2.ID)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, 1.5, body.Current.MedianEventsAttended)
	})

	t.Run("played-once count and share, and the 10+ cohort with exactly-10 counting", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		events := make([]int32, 10)
		for i := range events {
			e, err := createDashboardTestEvent(db, semester.ID, structure.ID, fmt.Sprintf("E%d", i), models.EventStateEnded, time.Now(), 0)
			require.NoError(t, err)
			events[i] = e.ID
		}

		// One played-once player, two 4-player group played-twice, and one player who
		// hit exactly 10 events.
		entrant(t, db, semester.ID, "once", events[0])
		entrant(t, db, semester.ID, "twice-a", events[0], events[1])
		entrant(t, db, semester.ID, "twice-b", events[0], events[1])
		entrant(t, db, semester.ID, "tenplus", events...)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 4, body.Current.Players)
		require.EqualValues(t, 1, body.Current.PlayedOnceCount)
		require.Equal(t, 0.25, body.Current.PlayedOnceShare)
		require.EqualValues(t, 1, body.Current.TenPlusCount)
	})

	t.Run("returns zeroes, not an error, for a semester with zero entries", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 0, body.Current.Players)
		require.Equal(t, 0.0, body.Current.MedianEventsAttended)
		require.EqualValues(t, 0, body.Current.PlayedOnceCount)
		require.Equal(t, 0.0, body.Current.PlayedOnceShare)
		require.EqualValues(t, 0, body.Current.TenPlusCount)
	})

	t.Run("returns comparison null when there is no comparable prior term", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
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
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		// Dated to the comparison term's own start: comparison stats are clipped to the
		// same elapsed point the current term has reached, so an event dated today —
		// ten months past Fall 2025's end — is correctly excluded.
		event, err := createDashboardTestEvent(db, comparisonSemester.ID, structure.ID, "E1", models.EventStateEnded, comparisonSemester.StartDate, 0)
		require.NoError(t, err)

		entrant(t, db, comparisonSemester.ID, "p1", event.ID)

		w := getEngagementStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		require.Equal(t, "Fall 2025", body.Comparison.Semester.Name)
		require.EqualValues(t, 1, body.Comparison.Stats.Players)
		require.EqualValues(t, 0, body.Current.Players)
	})

	t.Run("clips the comparison term to the same elapsed point the current term has reached", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()

		// Anchored to now so the assertion holds whenever the suite runs: the current
		// term is 20 days old, so the comparison term is counted for its first 20 days.
		now := time.Now().UTC()
		currentStart := now.AddDate(0, 0, -20)
		comparisonStart := currentStart.AddDate(-1, 0, 0)

		comparisonSemester, err := createDashboardTestSemester(db, "Prior", comparisonStart)
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Current", currentStart)
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		// Day 2 of the prior term — inside the elapsed window.
		early, err := createDashboardTestEvent(db, comparisonSemester.ID, structure.ID, "Early",
			models.EventStateEnded, comparisonStart.AddDate(0, 0, 2), 0)
		require.NoError(t, err)
		// Day 60 — the prior term reached it, but the current term has not.
		late, err := createDashboardTestEvent(db, comparisonSemester.ID, structure.ID, "Late",
			models.EventStateEnded, comparisonStart.AddDate(0, 0, 60), 0)
		require.NoError(t, err)

		entrant(t, db, comparisonSemester.ID, "early-player", early.ID)
		entrant(t, db, comparisonSemester.ID, "late-player", late.ID)

		w := getEngagementStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)

		// Only the day-2 player counts. Without clipping this would be 2, and a term
		// three weeks old would look like a collapse against a completed one.
		require.EqualValues(t, 1, body.Comparison.Stats.Players)
	})

	t.Run("counts the comparison term in full once the current term has outrun its length", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()

		now := time.Now().UTC()
		// createDashboardTestSemester spans three months; 200 days is well past that.
		currentStart := now.AddDate(0, 0, -200)
		comparisonStart := currentStart.AddDate(-1, 0, 0)

		comparisonSemester, err := createDashboardTestSemester(db, "Prior", comparisonStart)
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Current", currentStart)
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		early, err := createDashboardTestEvent(db, comparisonSemester.ID, structure.ID, "Early",
			models.EventStateEnded, comparisonStart.AddDate(0, 0, 2), 0)
		require.NoError(t, err)
		late, err := createDashboardTestEvent(db, comparisonSemester.ID, structure.ID, "Late",
			models.EventStateEnded, comparisonStart.AddDate(0, 0, 60), 0)
		require.NoError(t, err)

		entrant(t, db, comparisonSemester.ID, "early-player", early.ID)
		entrant(t, db, comparisonSemester.ID, "late-player", late.ID)

		w := getEngagementStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		require.EqualValues(t, 2, body.Comparison.Stats.Players)
	})

	t.Run("a participant row left with a NULL membership_id after its membership is deleted is excluded", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E1", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)

		orphanMembership := entrant(t, db, semester.ID, "orphan", event.ID)
		entrant(t, db, semester.ID, "survivor", event.ID)

		// Deleting the membership sets participants.membership_id to NULL at the DB
		// level (ON DELETE SET NULL) rather than removing the entry row.
		require.NoError(t, db.Unscoped().Delete(&models.Membership{}, "id = ?", orphanMembership.ID).Error)

		w := getEngagementStats(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.Players)
	})

	t.Run("a member entering via two different memberships in this semester's events counts once with combined events", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		priorSemester, err := createDashboardTestSemester(db, "Winter 2025", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		currentSemester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		// Both events belong to the current semester, but the entries are made
		// through two different memberships for the same user - one from a prior
		// semester, one from the current semester. Nothing in CreateParticipant
		// stops an entry from referencing a membership whose semester differs from
		// the event's semester.
		e1, err := createDashboardTestEvent(db, currentSemester.ID, structure.ID, "E1", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)
		e2, err := createDashboardTestEvent(db, currentSemester.ID, structure.ID, "E2", models.EventStateEnded, time.Now(), 0)
		require.NoError(t, err)

		user, err := testutils.CreateTestUser(db, nextUserID(), "Cross", "Term", "crossterm-eng@uwaterloo.ca", models.FacultyMath, "cte1")
		require.NoError(t, err)
		oldMembership, err := createDashboardTestMembership(db, user.ID, priorSemester.ID, true, false, false)
		require.NoError(t, err)
		newMembership, err := createDashboardTestMembership(db, user.ID, currentSemester.ID, true, false, false)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, oldMembership.ID, e1.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, newMembership.ID, e2.ID)
		require.NoError(t, err)

		w := getEngagementStats(t, currentSemester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EngagementStatsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.Players)
		require.Equal(t, 2.0, body.Current.MedianEventsAttended)
		require.EqualValues(t, 0, body.Current.PlayedOnceCount)
	})

	t.Run("returns 404 for an unknown semester id", func(t *testing.T) {
		w := getEngagementStats(t, "00000000-0000-0000-0000-000000000000")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for a malformed semester id", func(t *testing.T) {
		w := getEngagementStats(t, "not-a-uuid")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDashboardTrialConversion(t *testing.T) {
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
			fmt.Sprintf("/api/v2/semesters/%s/dashboard/conversion", semester.ID),
			[]string{"bot"},
		)
	})

	getConversion := func(t *testing.T, semesterID string) *httptest.ResponseRecorder {
		t.Helper()
		sessionID, err := testutils.CreateTestSession(db, "exec-"+uuid.NewString(), "executive")
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest(
			"GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/conversion", semesterID), nil,
		)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		return w
	}

	nextUserID := func() uint64 { return uint64(time.Now().UnixNano()) }
	createPlayer := func(t *testing.T, db *gorm.DB, semesterID uuid.UUID, tag string, paid, executive bool) *models.Membership {
		t.Helper()
		user, err := testutils.CreateTestUser(db, nextUserID(), tag, "Player", tag+"@uwaterloo.ca", models.FacultyMath, tag)
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, semesterID, paid, false, executive)
		require.NoError(t, err)
		return membership
	}
	setFreeTrialLimit := func(t *testing.T, db *gorm.DB, semesterID uuid.UUID, limit uint8) {
		t.Helper()
		require.NoError(t, db.Model(&models.Semester{}).Where("id = ?", semesterID).Update("free_trial_limit", limit).Error)
	}
	decode := func(t *testing.T, w *httptest.ResponseRecorder) controller.TrialConversionResponse {
		t.Helper()
		require.Equal(t, http.StatusOK, w.Code)
		var body controller.TrialConversionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		return body
	}

	t.Run("returns zeroes and no comparison when no eligible prior term exists", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Empty", time.Now().UTC())
		require.NoError(t, err)
		setFreeTrialLimit(t, db, semester.ID, 3)
		createPlayer(t, db, semester.ID, "bystander", true, false)

		body := decode(t, getConversion(t, semester.ID.String()))
		require.EqualValues(t, 3, body.FreeTrialLimit)
		require.Equal(t, store.TrialConversionStats{}, body.Current)
		require.EqualValues(t, 0, body.Conversion.Numerator)
		require.EqualValues(t, 0, body.Conversion.Denominator)
		require.Nil(t, body.Conversion.Rate)
		require.Nil(t, body.Comparison)
	})

	t.Run("calculates comparison status using the comparison term limit", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		comparison, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		current, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		setFreeTrialLimit(t, db, comparison.ID, 2)
		setFreeTrialLimit(t, db, current.ID, 3)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		comparisonE1, err := createDashboardTestEvent(db, comparison.ID, structure.ID, "Prior 1", models.EventStateEnded, comparison.StartDate, 0)
		require.NoError(t, err)
		comparisonE2, err := createDashboardTestEvent(db, comparison.ID, structure.ID, "Prior 2", models.EventStateEnded, comparison.StartDate.AddDate(0, 0, 7), 0)
		require.NoError(t, err)
		currentE1, err := createDashboardTestEvent(db, current.ID, structure.ID, "Current 1", models.EventStateEnded, current.StartDate, 0)
		require.NoError(t, err)
		currentE2, err := createDashboardTestEvent(db, current.ID, structure.ID, "Current 2", models.EventStateEnded, current.StartDate.AddDate(0, 0, 7), 0)
		require.NoError(t, err)

		priorPlayer := createPlayer(t, db, comparison.ID, "prior", false, false)
		currentPlayer := createPlayer(t, db, current.ID, "current", false, false)
		for _, entry := range []struct {
			membershipID uuid.UUID
			eventID      int32
		}{
			{priorPlayer.ID, comparisonE1.ID},
			{priorPlayer.ID, comparisonE2.ID},
			{currentPlayer.ID, currentE1.ID},
			{currentPlayer.ID, currentE2.ID},
		} {
			_, err = testutils.CreateTestParticipant(db, entry.membershipID, entry.eventID)
			require.NoError(t, err)
		}

		body := decode(t, getConversion(t, current.ID.String()))
		require.EqualValues(t, 3, body.FreeTrialLimit)
		require.EqualValues(t, 1, body.Current.TrialOpen)
		require.NotNil(t, body.Comparison)
		require.Equal(t, comparison.ID, body.Comparison.Semester.ID)
		require.Equal(t, comparison.Name, body.Comparison.Semester.Name)
		require.EqualValues(t, 2, body.Comparison.FreeTrialLimit)
		require.EqualValues(t, 1, body.Comparison.Stats.Players)
		require.EqualValues(t, 1, body.Comparison.Stats.TrialSpent)
		require.EqualValues(t, 0, body.Comparison.Stats.TrialOpen)
	})

	t.Run("reports a disabled trial for the comparison term", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		comparison, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		current, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		setFreeTrialLimit(t, db, current.ID, 3)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		comparisonEvent, err := createDashboardTestEvent(db, comparison.ID, structure.ID, "Prior", models.EventStateEnded, comparison.StartDate, 0)
		require.NoError(t, err)
		priorPlayer := createPlayer(t, db, comparison.ID, "prior-disabled", false, false)
		_, err = testutils.CreateTestParticipant(db, priorPlayer.ID, comparisonEvent.ID)
		require.NoError(t, err)

		body := decode(t, getConversion(t, current.ID.String()))
		require.NotNil(t, body.Comparison)
		require.EqualValues(t, 0, body.Comparison.FreeTrialLimit)
		// The persisted partition remains literal (entries >= 0), while clients use
		// the limit to present these unpaid players as not having had a trial.
		require.EqualValues(t, 1, body.Comparison.Stats.TrialSpent)
	})

	t.Run("limit zero preserves the literal spent-trial partition", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "No trial", time.Now().UTC())
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E1", models.EventStateEnded, time.Now().UTC(), 0)
		require.NoError(t, err)
		unpaid := createPlayer(t, db, semester.ID, "unpaid", false, false)
		_, err = testutils.CreateTestParticipant(db, unpaid.ID, event.ID)
		require.NoError(t, err)

		body := decode(t, getConversion(t, semester.ID.String()))
		require.EqualValues(t, 0, body.FreeTrialLimit)
		require.EqualValues(t, 1, body.Current.Players)
		require.EqualValues(t, 1, body.Current.TrialSpent)
		require.EqualValues(t, 0, body.Current.TrialOpen)
	})

	t.Run("partitions players by current status and distinct event entries", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Now().UTC())
		require.NoError(t, err)
		setFreeTrialLimit(t, db, semester.ID, 2)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		e1, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E1", models.EventStateEnded, time.Now().UTC(), 0)
		require.NoError(t, err)
		e2, err := createDashboardTestEvent(db, semester.ID, structure.ID, "E2", models.EventStateEnded, time.Now().UTC(), 0)
		require.NoError(t, err)

		paidAfterTrial := createPlayer(t, db, semester.ID, "paid", false, false)
		spent := createPlayer(t, db, semester.ID, "spent", false, false)
		open := createPlayer(t, db, semester.ID, "open", false, false)
		executive := createPlayer(t, db, semester.ID, "executive", false, true)
		for _, membership := range []*models.Membership{paidAfterTrial, spent} {
			_, err = testutils.CreateTestParticipant(db, membership.ID, e1.ID)
			require.NoError(t, err)
			_, err = testutils.CreateTestParticipant(db, membership.ID, e2.ID)
			require.NoError(t, err)
		}
		membershipRepo := postgresstore.NewMembershipRepository(db)
		require.NoError(t, membershipRepo.SetFreeTrialAvailable(paidAfterTrial.ID, false))
		paidAfterTrial.Paid = true
		require.NoError(t, membershipRepo.Update(paidAfterTrial))
		_, err = testutils.CreateTestParticipant(db, open.ID, e1.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, executive.ID, e1.ID)
		require.NoError(t, err)

		body := decode(t, getConversion(t, semester.ID.String()))
		require.EqualValues(t, 4, body.Current.Players)
		require.EqualValues(t, 1, body.Current.Paid)
		require.EqualValues(t, 1, body.Current.TrialSpent)
		require.EqualValues(t, 1, body.Current.TrialOpen)
		require.EqualValues(t, 1, body.Current.Executive)
		require.Equal(t, body.Current.Players, body.Current.Paid+body.Current.TrialSpent+body.Current.TrialOpen+body.Current.Executive)
	})

	t.Run("excludes cross-semester participant memberships", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		prior, err := createDashboardTestSemester(db, "Winter 2025", time.Now().UTC().AddDate(0, -4, 0))
		require.NoError(t, err)
		current, err := createDashboardTestSemester(db, "Fall 2025", time.Now().UTC())
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := createDashboardTestEvent(db, current.ID, structure.ID, "E1", models.EventStateEnded, time.Now().UTC(), 0)
		require.NoError(t, err)

		user, err := testutils.CreateTestUser(db, nextUserID(), "Cross", "Term", "cross-term-conversion@uwaterloo.ca", models.FacultyMath, "ctc1")
		require.NoError(t, err)
		oldMembership, err := createDashboardTestMembership(db, user.ID, prior.ID, true, false, false)
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, current.ID, false, false, false)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, oldMembership.ID, event.ID)
		require.NoError(t, err)

		body := decode(t, getConversion(t, current.ID.String()))
		require.Equal(t, store.TrialConversionStats{}, body.Current)
	})

	t.Run("reports observed cohorts separately from upfront, legacy, and comped entrants for both terms", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()
		comparison, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		current, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		setFreeTrialLimit(t, db, comparison.ID, 4)
		setFreeTrialLimit(t, db, current.ID, 4)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		priorEvent, err := createDashboardTestEvent(db, comparison.ID, structure.ID, "Prior", models.EventStateEnded, comparison.StartDate, 0)
		require.NoError(t, err)
		currentEvent, err := createDashboardTestEvent(db, current.ID, structure.ID, "Current", models.EventStateEnded, current.StartDate, 0)
		require.NoError(t, err)

		priorConverted := createPlayer(t, db, comparison.ID, "prior-converted", true, false)
		priorOpen := createPlayer(t, db, comparison.ID, "prior-open", false, false)
		priorUpfront := createPlayer(t, db, comparison.ID, "prior-upfront", true, false)
		priorUnknown := createPlayer(t, db, comparison.ID, "prior-legacy", false, false)
		priorExecutive := createPlayer(t, db, comparison.ID, "prior-exec", false, true)
		for _, membership := range []*models.Membership{priorConverted, priorOpen, priorUpfront, priorUnknown, priorExecutive} {
			_, err = testutils.CreateTestParticipant(db, membership.ID, priorEvent.ID)
			require.NoError(t, err)
		}
		startedAt := time.Date(2025, 9, 10, 18, 0, 0, 0, time.UTC)
		convertedAt := time.Date(2025, 9, 20, 18, 0, 0, 0, time.UTC)
		require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", priorConverted.ID).Updates(map[string]any{
			"trial_started_at": startedAt, "converted_at": convertedAt,
		}).Error)
		require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", priorOpen.ID).Update("trial_started_at", startedAt).Error)

		currentStarter := createPlayer(t, db, current.ID, "current-starter", false, false)
		_, err = testutils.CreateTestParticipant(db, currentStarter.ID, currentEvent.ID)
		require.NoError(t, err)
		require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", currentStarter.ID).Update("trial_started_at", startedAt).Error)
		// Deleting the live participant row must not erase the recorded trial cohort.
		require.NoError(t, postgresstore.NewEntryRepository(db).Delete(currentStarter.ID, currentEvent.ID))

		body := decode(t, getConversion(t, current.ID.String()))
		require.EqualValues(t, 0, body.Current.Players)
		require.EqualValues(t, 0, body.Conversion.Numerator)
		require.EqualValues(t, 1, body.Conversion.Denominator)
		require.NotNil(t, body.Conversion.Rate)
		require.InDelta(t, 0.0, *body.Conversion.Rate, 0.0001)
		require.NotNil(t, body.Comparison)
		require.EqualValues(t, 1, body.Comparison.Conversion.Numerator)
		require.EqualValues(t, 2, body.Comparison.Conversion.Denominator)
		require.InDelta(t, 0.5, *body.Comparison.Conversion.Rate, 0.0001)
		require.EqualValues(t, 2, body.Comparison.Conversion.UntrackedEntrants, "upfront buyer and historical unknown are context, not cohort members")
	})

	t.Run("returns 404 for an unknown semester and 400 for a malformed id", func(t *testing.T) {
		require.Equal(t, http.StatusNotFound, getConversion(t, "00000000-0000-0000-0000-000000000000").Code)
		require.Equal(t, http.StatusBadRequest, getConversion(t, "not-a-uuid").Code)
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

func TestDashboardEventActivity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)
	semester, err := testutils.CreateTestSemester(db, "Fall 2025")
	require.NoError(t, err)

	get := func(t *testing.T, semesterID string) *httptest.ResponseRecorder {
		sessionID, err := testutils.CreateTestSession(db, "event-activity-"+uuid.NewString(), "executive")
		require.NoError(t, err)
		req, err := testutils.MakeJSONRequest("GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/events", semesterID), nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		return w
	}
	newUserID := func() uint64 { return uint64(time.Now().UnixNano()) }

	t.Run("requires executive authorization", func(t *testing.T) {
		testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/events", semester.ID), []string{"bot"})
	})

	t.Run("counts ended events and entries while excluding started events", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		semester, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		full, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Full", models.EventStateEnded, time.Date(2025, 9, 5, 18, 0, 0, 0, time.UTC), 0)
		require.NoError(t, err)
		empty, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Empty", models.EventStateEnded, time.Date(2025, 9, 12, 18, 0, 0, 0, time.UTC), 0)
		require.NoError(t, err)
		scheduled, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Scheduled", models.EventStateStarted, time.Date(2025, 9, 19, 18, 0, 0, 0, time.UTC), 0)
		require.NoError(t, err)
		user, err := testutils.CreateTestUser(db, newUserID(), "Event", "Player", "event-player@uwaterloo.ca", models.FacultyMath, "event-player")
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
		require.NoError(t, err)
		secondUser, err := testutils.CreateTestUser(db, newUserID(), "Second", "Player", "second-event-player@uwaterloo.ca", models.FacultyMath, "second-event-player")
		require.NoError(t, err)
		secondMembership, err := createDashboardTestMembership(db, secondUser.ID, semester.ID, true, false, false)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership.ID, full.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, secondMembership.ID, full.ID)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership.ID, scheduled.ID)
		require.NoError(t, err)

		w := get(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		var body controller.EventActivityResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 2, body.Current.EventsRun)
		require.EqualValues(t, 1, body.Current.EventsScheduled)
		require.EqualValues(t, 2, body.Current.TotalEntries)
		require.Equal(t, 1.0, body.Current.AverageFieldSize)
		require.Equal(t, []int32{full.ID, empty.ID}, []int32{body.Current.Series[0].ID, body.Current.Series[1].ID})
		require.EqualValues(t, 0, body.Current.Series[1].Entries)
	})

	t.Run("a restarted event is scheduled again", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Restarted", models.EventStateEnded, time.Now().UTC(), 0)
		require.NoError(t, err)
		require.NoError(t, db.Model(event).Update("state", models.EventStateStarted).Error)
		var body controller.EventActivityResponse
		require.NoError(t, json.Unmarshal(get(t, semester.ID.String()).Body.Bytes(), &body))
		require.Zero(t, body.Current.EventsRun)
		require.EqualValues(t, 1, body.Current.EventsScheduled)
		require.NotNil(t, body.Current.Series)
		require.Empty(t, body.Current.Series)
	})

	t.Run("counts an entry whose membership was deleted", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		event, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Ended", models.EventStateEnded, time.Now().UTC(), 0)
		require.NoError(t, err)
		user, err := testutils.CreateTestUser(db, newUserID(), "Orphan", "Entry", "orphan-event@uwaterloo.ca", models.FacultyMath, "orphan-event")
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership.ID, event.ID)
		require.NoError(t, err)
		require.NoError(t, db.Unscoped().Delete(&models.Membership{}, "id = ?", membership.ID).Error)
		var body controller.EventActivityResponse
		require.NoError(t, json.Unmarshal(get(t, semester.ID.String()).Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.TotalEntries)
	})

	t.Run("orders equal start dates by id", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		date := time.Date(2025, 9, 5, 18, 0, 0, 0, time.UTC)
		first, err := createDashboardTestEvent(db, semester.ID, structure.ID, "First", models.EventStateEnded, date, 0)
		require.NoError(t, err)
		second, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Second", models.EventStateEnded, date, 0)
		require.NoError(t, err)
		later, err := createDashboardTestEvent(db, semester.ID, structure.ID, "Later", models.EventStateEnded, date.AddDate(0, 0, 7), 0)
		require.NoError(t, err)
		var body controller.EventActivityResponse
		require.NoError(t, json.Unmarshal(get(t, semester.ID.String()).Body.Bytes(), &body))
		require.Equal(t, []int32{first.ID, second.ID, later.ID}, []int32{body.Current.Series[0].ID, body.Current.Series[1].ID, body.Current.Series[2].ID})
	})

	t.Run("returns zeroes, an empty series, and null comparison without data", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		semester, err := testutils.CreateTestSemester(db, "Fall 2025")
		require.NoError(t, err)
		w := get(t, semester.ID.String())
		var body controller.EventActivityResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, controller.EventActivityCurrent{EventActivityStats: store.EventActivityStats{}, Series: []store.EventSeriesPoint{}}, body.Current)
		require.Nil(t, body.Comparison)
		require.Contains(t, w.Body.String(), "\"comparison\":null")
	})

	t.Run("returns only the comparison average and semester reference", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		prior, err := createDashboardTestSemester(db, "Fall 2025", time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		current, err := createDashboardTestSemester(db, "Fall 2026", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)
		// See the engagement suite: comparison figures are clipped to the elapsed point,
		// so the event must sit inside the term it belongs to.
		event, err := createDashboardTestEvent(db, prior.ID, structure.ID, "Prior", models.EventStateEnded, prior.StartDate, 0)
		require.NoError(t, err)
		user, err := testutils.CreateTestUser(db, newUserID(), "Prior", "Player", "prior-player@uwaterloo.ca", models.FacultyMath, "prior-player")
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, prior.ID, true, false, false)
		require.NoError(t, err)
		_, err = testutils.CreateTestParticipant(db, membership.ID, event.ID)
		require.NoError(t, err)
		w := get(t, current.ID.String())
		var raw map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
		comparison := raw["comparison"].(map[string]any)
		require.Equal(t, map[string]any{"semester": comparison["semester"], "averageFieldSize": comparison["averageFieldSize"]}, comparison)
		require.Equal(t, prior.ID.String(), comparison["semester"].(map[string]any)["id"])
		require.Equal(t, 1.0, comparison["averageFieldSize"])
	})

	t.Run("counts future-dated events as scheduled rather than clipping them away", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db := container.GetDB()

		now := time.Now().UTC()
		semester, err := createDashboardTestSemester(db, "Current", now.AddDate(0, 0, -20))
		require.NoError(t, err)
		structure, err := testutils.CreateTestStructure(db, "Standard")
		require.NoError(t, err)

		_, err = createDashboardTestEvent(db, semester.ID, structure.ID, "Ran",
			models.EventStateEnded, now.AddDate(0, 0, -7), 0)
		require.NoError(t, err)
		// Next week's event. A scheduled event is future-dated by definition, so the
		// asOf cutoff must not reach it.
		_, err = createDashboardTestEvent(db, semester.ID, structure.ID, "Upcoming",
			models.EventStateStarted, now.AddDate(0, 0, 7), 0)
		require.NoError(t, err)

		w := get(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)

		var body controller.EventActivityResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.EqualValues(t, 1, body.Current.EventsRun)
		require.EqualValues(t, 1, body.Current.EventsScheduled)
	})

	t.Run("returns 400 for malformed and 404 for unknown semester ids", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, get(t, "not-a-uuid").Code)
		require.Equal(t, http.StatusNotFound, get(t, "00000000-0000-0000-0000-000000000000").Code)
	})
}

func TestDashboardSignupTimeline(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)
	semester, err := testutils.CreateTestSemester(db, "Fall 2025")
	require.NoError(t, err)
	toronto, err := time.LoadLocation("America/Toronto")
	require.NoError(t, err)

	get := func(t *testing.T, semesterID string) *httptest.ResponseRecorder {
		sessionID, err := testutils.CreateTestSession(db, "signup-timeline-"+uuid.NewString(), "executive")
		require.NoError(t, err)
		req, err := testutils.MakeJSONRequest("GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/signups", semesterID), nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		return w
	}
	newUserID := func() uint64 { return uint64(time.Now().UnixNano()) }

	t.Run("requires executive authorization", func(t *testing.T) {
		testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "GET", fmt.Sprintf("/api/v2/semesters/%s/dashboard/signups", semester.ID), []string{"bot"})
	})

	t.Run("returns literal calendar dates, zero-filled source buckets, and matching total", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		today := time.Now().In(toronto)
		day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, toronto)
		semester, err := createDashboardTestSemester(db, "Current", day.AddDate(0, 0, -2))
		require.NoError(t, err)

		addMembership := func(source *models.MembershipSource, createdAt *time.Time) {
			user, err := testutils.CreateTestUser(db, newUserID(), "Signup", uuid.NewString(), uuid.NewString()+"@uwaterloo.ca", models.FacultyMath, "signup")
			require.NoError(t, err)
			membership, err := createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
			require.NoError(t, err)
			require.NoError(t, db.Model(membership).Updates(map[string]any{"source": source, "created_at": createdAt}).Error)
		}

		admin := models.MembershipSourceAdmin
		discord := models.MembershipSourceDiscord
		malformed := models.MembershipSource("kiosk")
		first := day.Add(9 * time.Hour)
		last := day.AddDate(0, 0, 2).Add(20 * time.Hour)
		addMembership(&admin, &first)
		addMembership(&discord, &first)
		addMembership(nil, &last)
		addMembership(&malformed, &last)
		addMembership(&admin, nil) // A never-participating historical membership remains undated.

		structure, err := testutils.CreateTestStructure(db, "Signup timeline")
		require.NoError(t, err)
		_, err = createDashboardTestEvent(db, semester.ID, structure.ID, "In term", models.EventStateStarted, day.AddDate(0, 0, 1).Add(18*time.Hour), 0)
		require.NoError(t, err)
		other, err := createDashboardTestSemester(db, "Other", day)
		require.NoError(t, err)
		_, err = createDashboardTestEvent(db, other.ID, structure.ID, "Other term", models.EventStateStarted, day.Add(18*time.Hour), 0)
		require.NoError(t, err)

		w := get(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		var body store.SignupTimeline
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, []store.SignupTimelinePoint{
			{Date: day.AddDate(0, 0, -2).Format("2006-01-02")},
			{Date: day.AddDate(0, 0, -1).Format("2006-01-02")},
			{Date: day.Format("2006-01-02"), Admin: 1, Discord: 1},
		}, body.Series)
		require.Equal(t, []string{day.AddDate(0, 0, 1).Format("2006-01-02")}, body.EventDates)
		require.NotNil(t, body.DataStartsAt)
		require.Equal(t, day.Format("2006-01-02"), *body.DataStartsAt)
		require.EqualValues(t, 2, body.Total)
		var bucketSum int64
		for _, point := range body.Series {
			bucketSum += point.Admin + point.Discord + point.Unknown
			require.Len(t, point.Date, len("2006-01-02"))
			require.NotContains(t, point.Date, "T")
		}
		require.Equal(t, body.Total, bucketSum)
		require.Contains(t, w.Body.String(), `"dataStartsAt":"`+day.Format("2006-01-02")+`"`)
	})

	t.Run("uses unknown for null and malformed sources within the current bound", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		today := time.Now().In(toronto)
		day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, toronto)
		semester, err := createDashboardTestSemester(db, "Current", day)
		require.NoError(t, err)
		for _, source := range []*models.MembershipSource{nil, func() *models.MembershipSource { value := models.MembershipSource("future"); return &value }()} {
			user, err := testutils.CreateTestUser(db, newUserID(), "Unknown", uuid.NewString(), uuid.NewString()+"@uwaterloo.ca", models.FacultyMath, "unknown")
			require.NoError(t, err)
			membership, err := createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
			require.NoError(t, err)
			require.NoError(t, db.Model(membership).Updates(map[string]any{"source": source, "created_at": day.Add(12 * time.Hour)}).Error)
		}
		var body store.SignupTimeline
		require.NoError(t, json.Unmarshal(get(t, semester.ID.String()).Body.Bytes(), &body))
		require.Len(t, body.Series, 1)
		require.EqualValues(t, 2, body.Series[0].Unknown)
		require.EqualValues(t, 2, body.Total)
		require.Equal(t, body.Total, body.Series[0].Admin+body.Series[0].Discord+body.Series[0].Unknown)
	})

	t.Run("returns an empty non-nil series when the current bound precedes the semester", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		future := time.Now().In(toronto).AddDate(0, 0, 2)
		semester, err := createDashboardTestSemester(db, "Future", future)
		require.NoError(t, err)
		var body store.SignupTimeline
		w := get(t, semester.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Series)
		require.Empty(t, body.Series)
		require.NotNil(t, body.EventDates)
		require.Nil(t, body.DataStartsAt)
		require.Zero(t, body.Total)
	})

	t.Run("uses the Toronto calendar date for the current bound", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		semester, err := createDashboardTestSemester(db, "Toronto boundary", time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		user, err := testutils.CreateTestUser(db, newUserID(), "Toronto", "Boundary", "toronto-boundary@uwaterloo.ca", models.FacultyMath, "toronto")
		require.NoError(t, err)
		membership, err := createDashboardTestMembership(db, user.ID, semester.ID, true, false, false)
		require.NoError(t, err)
		createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		require.NoError(t, db.Model(membership).Update("created_at", createdAt).Error)

		// 02:00 UTC on January 1 is still December 31 in Toronto. Passing this
		// instant directly to the repository makes the SQL time-zone contract
		// deterministic instead of depending on the test runner's clock.
		timeline, err := postgresstore.NewDashboardRepository(db).SignupTimeline(
			semester.ID, time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC),
		)
		require.NoError(t, err)
		require.Equal(t, []store.SignupTimelinePoint{
			{Date: "2025-12-30"},
			{Date: "2025-12-31"},
		}, timeline.Series)
		require.Nil(t, timeline.DataStartsAt)
		require.Zero(t, timeline.Total)
	})

	t.Run("returns previous same-season daily totals aligned by calendar-day offset and clipped to term coverage", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		today := time.Now().In(toronto)
		currentStart := time.Date(today.Year(), today.Month(), today.Day()-4, 0, 0, 0, 0, toronto)
		current, err := createDashboardTestSemester(db, "Fall current", currentStart)
		require.NoError(t, err)
		comparisonStart := time.Date(currentStart.Year()-1, time.September, 3, 0, 0, 0, 0, toronto)
		comparison, err := createDashboardTestSemester(db, "Fall prior", comparisonStart)
		require.NoError(t, err)
		require.NoError(t, db.Model(comparison).Update("end_date", comparisonStart.AddDate(0, 0, 2).Add(12*time.Hour)).Error)

		addDated := func(semesterID uuid.UUID, createdAt *time.Time, source *models.MembershipSource) {
			user, err := testutils.CreateTestUser(db, newUserID(), "Comparison", uuid.NewString(), uuid.NewString()+"@uwaterloo.ca", models.FacultyMath, "comparison")
			require.NoError(t, err)
			membership, err := createDashboardTestMembership(db, user.ID, semesterID, true, false, false)
			require.NoError(t, err)
			require.NoError(t, db.Model(membership).Updates(map[string]any{"created_at": createdAt, "source": source}).Error)
		}
		admin := models.MembershipSourceAdmin
		discord := models.MembershipSourceDiscord
		firstDay := comparisonStart.Add(10 * time.Hour)
		thirdDay := comparisonStart.AddDate(0, 0, 2).Add(10 * time.Hour)
		addDated(comparison.ID, &firstDay, &admin)
		addDated(comparison.ID, &firstDay, &discord)
		addDated(comparison.ID, &thirdDay, &admin)

		w := get(t, current.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		var body controller.SignupTimelineResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotNil(t, body.Comparison)
		require.Equal(t, comparison.ID, body.Comparison.Semester.ID)
		require.Equal(t, "Fall prior", body.Comparison.Semester.Name)
		require.Equal(t, []controller.SignupTimelineComparisonPoint{
			{ElapsedDay: 0, Total: 2},
			{ElapsedDay: 1, Total: 0},
			{ElapsedDay: 2, Total: 1},
		}, body.Comparison.DailyTotals)
	})

	t.Run("does not invent comparison zeros when prior memberships have no dated records", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))
		db = container.GetDB()
		today := time.Now().In(toronto)
		currentStart := time.Date(today.Year(), today.Month(), today.Day()-2, 0, 0, 0, 0, toronto)
		current, err := createDashboardTestSemester(db, "Fall current", currentStart)
		require.NoError(t, err)
		priorStart := time.Date(currentStart.Year()-1, time.September, 3, 0, 0, 0, 0, toronto)
		prior, err := createDashboardTestSemester(db, "Fall prior undated", priorStart)
		require.NoError(t, err)
		user, err := testutils.CreateTestUser(db, newUserID(), "Undated", "Prior", uuid.NewString()+"@uwaterloo.ca", models.FacultyMath, "undated")
		require.NoError(t, err)
		_, err = createDashboardTestMembership(db, user.ID, prior.ID, true, false, false)
		require.NoError(t, err)
		currentUser, err := testutils.CreateTestUser(db, newUserID(), "Current", "Dated", uuid.NewString()+"@uwaterloo.ca", models.FacultyMath, "dated")
		require.NoError(t, err)
		currentMembership, err := createDashboardTestMembership(db, currentUser.ID, current.ID, true, false, false)
		require.NoError(t, err)
		createdAt := currentStart.Add(time.Hour)
		require.NoError(t, db.Model(currentMembership).Update("created_at", createdAt).Error)

		w := get(t, current.ID.String())
		require.Equal(t, http.StatusOK, w.Code)
		var body controller.SignupTimelineResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Nil(t, body.Comparison)
		require.Len(t, body.Series, 3)
	})

	t.Run("returns 400 for malformed and 404 for unknown semester ids", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, get(t, "not-a-uuid").Code)
		require.Equal(t, http.StatusNotFound, get(t, "00000000-0000-0000-0000-000000000000").Code)
	})
}
