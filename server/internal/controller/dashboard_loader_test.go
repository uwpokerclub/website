package controller

import (
	"api/internal/models"
	"api/internal/store"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type dashboardLoaderTestStore struct {
	store.Store
	semesters *dashboardLoaderTestSemesterRepository
	dashboard *dashboardLoaderTestDashboardRepository
}

func (s dashboardLoaderTestStore) Semesters() store.SemesterRepository  { return s.semesters }
func (s dashboardLoaderTestStore) Dashboard() store.DashboardRepository { return s.dashboard }

type dashboardLoaderTestSemesterRepository struct {
	store.SemesterRepository
	semester  models.Semester
	semesters []models.Semester
	findErr   error
	listErr   error
	findCalls int
	listCalls int
}

func (r *dashboardLoaderTestSemesterRepository) FindByID(uuid.UUID) (models.Semester, error) {
	r.findCalls++
	return r.semester, r.findErr
}

func (r *dashboardLoaderTestSemesterRepository) List(*models.Pagination) ([]models.Semester, int64, error) {
	r.listCalls++
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	if r.semesters != nil {
		return r.semesters, int64(len(r.semesters)), nil
	}
	return []models.Semester{r.semester}, 1, nil
}

type dashboardLoaderTestDashboardRepository struct {
	store.DashboardRepository
	timeline  store.SignupTimeline
	stats     map[uuid.UUID]store.MembershipStats
	partial   *int64
	asOfCalls int
}

func (r *dashboardLoaderTestDashboardRepository) Spotlight(uuid.UUID, time.Time) (*store.SpotlightEvent, error) {
	return nil, nil
}

// Fall back to zeroes for the loader tests that do not exercise membership stats.
func (r *dashboardLoaderTestDashboardRepository) MembershipStats(semesterID uuid.UUID) (store.MembershipStats, error) {
	if r.stats != nil {
		return r.stats[semesterID], nil
	}
	return store.MembershipStats{}, nil
}
func (r *dashboardLoaderTestDashboardRepository) EngagementStats(uuid.UUID, time.Time) (store.EngagementStats, error) {
	return store.EngagementStats{}, nil
}
func (r *dashboardLoaderTestDashboardRepository) TrialConversionStats(uuid.UUID, uint8) (store.TrialConversionStats, error) {
	return store.TrialConversionStats{}, nil
}
func (r *dashboardLoaderTestDashboardRepository) TrialConversionCohortStats(uuid.UUID) (store.TrialConversionCohortStats, error) {
	return store.TrialConversionCohortStats{}, nil
}
func (r *dashboardLoaderTestDashboardRepository) EventActivity(uuid.UUID, time.Time) (store.EventActivityStats, []store.EventSeriesPoint, error) {
	return store.EventActivityStats{}, []store.EventSeriesPoint{}, nil
}
func (r *dashboardLoaderTestDashboardRepository) AverageFieldSize(uuid.UUID, time.Time) (*float64, error) {
	return nil, nil
}
func (r *dashboardLoaderTestDashboardRepository) SignupTimeline(uuid.UUID, time.Time) (store.SignupTimeline, error) {
	return r.timeline, nil
}
func (r *dashboardLoaderTestDashboardRepository) MembershipTotalAsOf(uuid.UUID, time.Time) (*int64, error) {
	r.asOfCalls++
	return r.partial, nil
}

func newDashboardLoaderTestController(timeline store.SignupTimeline) (*dashboardController, *dashboardLoaderTestSemesterRepository) {
	semester := models.Semester{ID: uuid.New(), Name: "Fall 2025"}
	semesters := &dashboardLoaderTestSemesterRepository{semester: semester}
	return &dashboardController{
		store: dashboardLoaderTestStore{
			semesters: semesters,
			dashboard: &dashboardLoaderTestDashboardRepository{timeline: timeline},
		},
	}, semesters
}

func dashboardLoaderRequest(handler func(*gin.Context), semesterID string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "semesterId", Value: semesterID}}
	handler(ctx)
	return recorder
}

func TestDashboardSemesterLookupFailurePolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lookupErr := errors.New("injected semester lookup failure")
	tests := []struct {
		name     string
		handler  func(*dashboardController, *gin.Context)
		wantBody string
		wantLog  string
	}{
		{"spotlight", (*dashboardController).getSpotlight, lookupErr.Error(), ""},
		{"memberships", (*dashboardController).getMembershipStats, lookupErr.Error(), ""},
		{"engagement", (*dashboardController).getEngagementStats, lookupErr.Error(), ""},
		{"conversion", (*dashboardController).getTrialConversion, lookupErr.Error(), ""},
		{"events", (*dashboardController).getEventActivity, lookupErr.Error(), ""},
		{"signups", (*dashboardController).getSignupTimeline, "unable to load signup timeline", "dashboard signup timeline semester lookup failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
			semesters.findErr = lookupErr
			var output bytes.Buffer
			previousOutput := log.Writer()
			log.SetOutput(&output)
			t.Cleanup(func() { log.SetOutput(previousOutput) })

			recorder := dashboardLoaderRequest(func(ctx *gin.Context) { test.handler(controller, ctx) }, semesters.semester.ID.String())
			require.Equal(t, http.StatusInternalServerError, recorder.Code)
			require.Contains(t, recorder.Body.String(), test.wantBody)
			if test.wantLog != "" {
				require.Contains(t, output.String(), test.wantLog)
				require.Contains(t, output.String(), lookupErr.Error())
			} else {
				require.Empty(t, output.String())
			}
			require.Equal(t, 1, semesters.findCalls)
			require.Zero(t, semesters.listCalls)
		})
	}
}

func TestMembershipComparisonUsesInclusiveTorontoCompletionBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loc, err := time.LoadLocation("America/Toronto")
	require.NoError(t, err)

	for _, test := range []struct {
		name    string
		endDate time.Time
	}{
		{name: "last UTC date label", endDate: time.Date(2025, time.April, 30, 0, 0, 0, 0, time.UTC)},
		{name: "DST transition", endDate: time.Date(2025, time.March, 8, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(test.name, func(t *testing.T) {
			comparison := models.Semester{ID: uuid.New(), Name: "Prior", StartDate: time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC), EndDate: test.endDate}
			targetEnd := time.Date(test.endDate.Year()+1, test.endDate.Month(), test.endDate.Day(), 0, 0, 0, 0, time.UTC)
			target := models.Semester{ID: uuid.New(), Name: "Current", StartDate: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC), EndDate: targetEnd}
			partial := int64(85)
			dashboard := &dashboardLoaderTestDashboardRepository{
				stats: map[uuid.UUID]store.MembershipStats{
					comparison.ID: {Total: 100},
					target.ID:     {},
				},
				partial: &partial,
			}
			semesters := &dashboardLoaderTestSemesterRepository{semester: target, semesters: []models.Semester{comparison, target}}
			now := time.Date(targetEnd.Year(), targetEnd.Month(), targetEnd.Day(), 23, 30, 0, 0, loc)
			controller := &dashboardController{
				store: dashboardLoaderTestStore{semesters: semesters, dashboard: dashboard},
				now:   func() time.Time { return now },
			}
			request := func() MembershipStatsResponse {
				recorder := dashboardLoaderRequest(controller.getMembershipStats, target.ID.String())
				require.Equal(t, http.StatusOK, recorder.Code)
				var response MembershipStatsResponse
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				return response
			}

			response := request()
			require.NotNil(t, response.Comparison)
			require.NotNil(t, response.Comparison.TotalAsOf)
			require.EqualValues(t, 85, *response.Comparison.TotalAsOf, "the UTC date label is not the exclusive Toronto boundary")
			require.Equal(t, 1, dashboard.asOfCalls)

			now = time.Date(targetEnd.Year(), targetEnd.Month(), targetEnd.Day()+1, 0, 0, 0, 0, loc)
			response = request()
			require.NotNil(t, response.Comparison)
			require.NotNil(t, response.Comparison.TotalAsOf)
			require.EqualValues(t, 100, *response.Comparison.TotalAsOf, "completed comparisons include undated membership rows")
			require.Equal(t, 1, dashboard.asOfCalls, "completed comparisons should use the exact aggregate")
		})
	}
}

func TestDashboardComparisonListFailurePolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	listErr := errors.New("injected semester list failure")
	tests := []struct {
		name     string
		handler  func(*dashboardController, *gin.Context)
		wantCode int
		wantBody string
		wantLog  string
	}{
		{"memberships", (*dashboardController).getMembershipStats, http.StatusInternalServerError, listErr.Error(), ""},
		{"engagement", (*dashboardController).getEngagementStats, http.StatusInternalServerError, listErr.Error(), ""},
		{"conversion", (*dashboardController).getTrialConversion, http.StatusInternalServerError, listErr.Error(), ""},
		{"events", (*dashboardController).getEventActivity, http.StatusInternalServerError, listErr.Error(), ""},
		{"signups", (*dashboardController).getSignupTimeline, http.StatusOK, `"comparison":null`, "dashboard signup comparison semester lookup failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timeline := store.SignupTimeline{Series: []store.SignupTimelinePoint{{Date: "2025-09-01"}}}
			controller, semesters := newDashboardLoaderTestController(timeline)
			semesters.listErr = listErr
			var output bytes.Buffer
			previousOutput := log.Writer()
			log.SetOutput(&output)
			t.Cleanup(func() { log.SetOutput(previousOutput) })

			recorder := dashboardLoaderRequest(func(ctx *gin.Context) { test.handler(controller, ctx) }, semesters.semester.ID.String())
			require.Equal(t, test.wantCode, recorder.Code)
			require.Contains(t, recorder.Body.String(), test.wantBody)
			if test.wantLog != "" {
				require.Contains(t, output.String(), test.wantLog)
				require.Contains(t, output.String(), listErr.Error())
			} else {
				require.Empty(t, output.String())
			}
			require.Equal(t, 1, semesters.findCalls)
			require.Equal(t, 1, semesters.listCalls)
		})
	}
}

func TestDashboardRoutesWithoutComparisonList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("spotlight never lists semesters", func(t *testing.T) {
		controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
		recorder := dashboardLoaderRequest(controller.getSpotlight, semesters.semester.ID.String())
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, 1, semesters.findCalls)
		require.Zero(t, semesters.listCalls)
	})
	t.Run("empty signup timeline skips comparison list", func(t *testing.T) {
		controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
		recorder := dashboardLoaderRequest(controller.getSignupTimeline, semesters.semester.ID.String())
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), `"comparison":null`)
		require.Equal(t, 1, semesters.findCalls)
		require.Zero(t, semesters.listCalls)
	})
}

func TestDashboardSemesterLoaderValidationAndNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		semesterID string
		findErr    error
		wantCode   int
	}{
		{"malformed id", "not-a-uuid", nil, http.StatusBadRequest},
		{"unknown id", uuid.NewString(), store.ErrNotFound, http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
			semesters.findErr = test.findErr
			recorder := dashboardLoaderRequest(controller.getSpotlight, test.semesterID)
			require.Equal(t, test.wantCode, recorder.Code)
			if test.name == "malformed id" {
				require.Zero(t, semesters.findCalls)
			} else {
				require.Equal(t, 1, semesters.findCalls)
			}
			require.Zero(t, semesters.listCalls)
		})
	}
}

func TestDashboardSignupLookupSanitizesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
	semesters.findErr = errors.New("database password and host details")
	recorder := dashboardLoaderRequest(controller.getSignupTimeline, semesters.semester.ID.String())
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), "unable to load signup timeline")
	require.NotContains(t, recorder.Body.String(), "database password")
}
