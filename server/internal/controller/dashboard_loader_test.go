package controller

import (
	"api/internal/models"
	"api/internal/store"
	"bytes"
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
	return []models.Semester{r.semester}, 1, nil
}

type dashboardLoaderTestDashboardRepository struct {
	store.DashboardRepository
	timeline store.SignupTimeline
}

func (r *dashboardLoaderTestDashboardRepository) Spotlight(uuid.UUID, time.Time) (*store.SpotlightEvent, error) {
	return nil, nil
}
func (r *dashboardLoaderTestDashboardRepository) MembershipStats(uuid.UUID) (store.MembershipStats, error) {
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
	return nil, nil
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
