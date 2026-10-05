package controller

import (
	"api/internal/models"
	"api/internal/services"
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
	timeline       store.SignupTimeline
	stats          map[uuid.UUID]store.MembershipStats
	partial        *int64
	asOfCalls      int
	asOfCutoffs    []time.Time
	spotlightNows  []time.Time
	engagementNows []time.Time
	eventNows      []time.Time
	averageNows    []time.Time
	signupNows     []time.Time
	spotlightErr   error
	membershipErr  error
	asOfErr        error
	engagementErr  error
	conversionErr  error
	cohortErr      error
	eventErr       error
	averageErr     error
	signupErr      error
}

func (r *dashboardLoaderTestDashboardRepository) Spotlight(_ uuid.UUID, now time.Time) (*store.SpotlightEvent, error) {
	r.spotlightNows = append(r.spotlightNows, now)
	return nil, r.spotlightErr
}

// Fall back to zeroes for the loader tests that do not exercise membership stats.
func (r *dashboardLoaderTestDashboardRepository) MembershipStats(semesterID uuid.UUID) (store.MembershipStats, error) {
	if r.membershipErr != nil {
		return store.MembershipStats{}, r.membershipErr
	}
	if r.stats != nil {
		return r.stats[semesterID], nil
	}
	return store.MembershipStats{}, nil
}
func (r *dashboardLoaderTestDashboardRepository) EngagementStats(_ uuid.UUID, now time.Time) (store.EngagementStats, error) {
	r.engagementNows = append(r.engagementNows, now)
	return store.EngagementStats{}, r.engagementErr
}
func (r *dashboardLoaderTestDashboardRepository) TrialConversionStats(uuid.UUID, uint8) (store.TrialConversionStats, error) {
	return store.TrialConversionStats{}, r.conversionErr
}
func (r *dashboardLoaderTestDashboardRepository) TrialConversionCohortStats(uuid.UUID) (store.TrialConversionCohortStats, error) {
	return store.TrialConversionCohortStats{}, r.cohortErr
}
func (r *dashboardLoaderTestDashboardRepository) EventActivity(_ uuid.UUID, now time.Time) (store.EventActivityStats, []store.EventSeriesPoint, error) {
	r.eventNows = append(r.eventNows, now)
	return store.EventActivityStats{}, []store.EventSeriesPoint{}, r.eventErr
}
func (r *dashboardLoaderTestDashboardRepository) AverageFieldSize(_ uuid.UUID, now time.Time) (*float64, error) {
	r.averageNows = append(r.averageNows, now)
	return nil, r.averageErr
}
func (r *dashboardLoaderTestDashboardRepository) SignupTimeline(_ uuid.UUID, now time.Time) (store.SignupTimeline, error) {
	r.signupNows = append(r.signupNows, now)
	return r.timeline, r.signupErr
}
func (r *dashboardLoaderTestDashboardRepository) MembershipTotalAsOf(_ uuid.UUID, cutoff time.Time) (*int64, error) {
	r.asOfCalls++
	r.asOfCutoffs = append(r.asOfCutoffs, cutoff)
	return r.partial, r.asOfErr
}

func newDashboardLoaderTestController(timeline store.SignupTimeline) (*dashboardController, *dashboardLoaderTestSemesterRepository) {
	semester := models.Semester{ID: uuid.New(), Name: "Fall 2025"}
	semesters := &dashboardLoaderTestSemesterRepository{semester: semester}
	return &dashboardController{
		store: dashboardLoaderTestStore{
			semesters: semesters,
			dashboard: &dashboardLoaderTestDashboardRepository{timeline: timeline},
		},
		now: func() time.Time { return time.Date(2026, time.January, 20, 12, 0, 0, 0, time.UTC) },
	}, semesters
}

func dashboardLoaderRequest(handler func(*gin.Context), semesterID string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "semesterId", Value: semesterID}}
	handler(ctx)
	return recorder
}

func TestDashboardSemesterLookupFailuresAreSanitizedAndLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lookupErr := errors.New("injected semester lookup failure")
	tests := []struct {
		name    string
		handler func(*dashboardController, *gin.Context)
	}{
		{"spotlight", (*dashboardController).getSpotlight},
		{"memberships", (*dashboardController).getMembershipStats},
		{"engagement", (*dashboardController).getEngagementStats},
		{"conversion", (*dashboardController).getTrialConversion},
		{"events", (*dashboardController).getEventActivity},
		{"signups", (*dashboardController).getSignupTimeline},
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
			require.Contains(t, recorder.Body.String(), "unable to load dashboard data")
			require.NotContains(t, recorder.Body.String(), lookupErr.Error())
			require.Contains(t, output.String(), "dashboard semester lookup failed")
			require.Contains(t, output.String(), lookupErr.Error())
			require.Equal(t, 1, semesters.findCalls)
			require.Zero(t, semesters.listCalls)
		})
	}
}

func TestDashboardAggregateFailuresAreSanitizedAndLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	queryErr := errors.New("database password and internal host")
	tests := []struct {
		name    string
		handler func(*dashboardController, *gin.Context)
		setErr  func(*dashboardLoaderTestDashboardRepository)
	}{
		{"spotlight", (*dashboardController).getSpotlight, func(repo *dashboardLoaderTestDashboardRepository) { repo.spotlightErr = queryErr }},
		{"memberships", (*dashboardController).getMembershipStats, func(repo *dashboardLoaderTestDashboardRepository) { repo.membershipErr = queryErr }},
		{"engagement", (*dashboardController).getEngagementStats, func(repo *dashboardLoaderTestDashboardRepository) { repo.engagementErr = queryErr }},
		{"conversion", (*dashboardController).getTrialConversion, func(repo *dashboardLoaderTestDashboardRepository) { repo.conversionErr = queryErr }},
		{"conversion cohort", (*dashboardController).getTrialConversion, func(repo *dashboardLoaderTestDashboardRepository) { repo.cohortErr = queryErr }},
		{"events", (*dashboardController).getEventActivity, func(repo *dashboardLoaderTestDashboardRepository) { repo.eventErr = queryErr }},
		{"signups", (*dashboardController).getSignupTimeline, func(repo *dashboardLoaderTestDashboardRepository) { repo.signupErr = queryErr }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
			repo := controller.store.(dashboardLoaderTestStore).dashboard
			test.setErr(repo)
			var output bytes.Buffer
			previousOutput := log.Writer()
			log.SetOutput(&output)
			t.Cleanup(func() { log.SetOutput(previousOutput) })

			recorder := dashboardLoaderRequest(func(ctx *gin.Context) { test.handler(controller, ctx) }, semesters.semester.ID.String())

			require.Equal(t, http.StatusInternalServerError, recorder.Code)
			require.Contains(t, recorder.Body.String(), "unable to load dashboard data")
			require.NotContains(t, recorder.Body.String(), "database password")
			require.NotContains(t, recorder.Body.String(), "internal host")
			require.Contains(t, output.String(), queryErr.Error())
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

func TestDashboardHandlersUseOneClockInstantAtTorontoEndBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	location, err := time.LoadLocation("America/Toronto")
	require.NoError(t, err)

	comparison := models.Semester{
		ID:        uuid.New(),
		Name:      "Prior Spring",
		StartDate: time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, time.April, 30, 0, 0, 0, 0, time.UTC),
	}
	target := models.Semester{
		ID:        uuid.New(),
		Name:      "Current Spring",
		StartDate: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC),
	}
	date := "2026-03-08"
	timeline := store.SignupTimeline{
		Series:       []store.SignupTimelinePoint{{Date: date, Admin: 1}},
		DataStartsAt: &date,
		Total:        1,
	}

	for _, test := range []struct {
		name string
		now  time.Time
	}{
		{name: "last Toronto day across spring DST", now: time.Date(2026, time.March, 8, 23, 59, 59, 0, location)},
		{name: "exclusive Toronto midnight after spring DST", now: time.Date(2026, time.March, 9, 0, 0, 0, 0, location)},
	} {
		t.Run(test.name, func(t *testing.T) {
			newController := func() (*dashboardController, *dashboardLoaderTestDashboardRepository) {
				dashboard := &dashboardLoaderTestDashboardRepository{
					timeline: timeline,
					stats: map[uuid.UUID]store.MembershipStats{
						target.ID:     {},
						comparison.ID: {Total: 100},
					},
				}
				semesters := &dashboardLoaderTestSemesterRepository{semester: target, semesters: []models.Semester{target, comparison}}
				return &dashboardController{
					store: dashboardLoaderTestStore{semesters: semesters, dashboard: dashboard},
					now:   func() time.Time { return test.now },
				}, dashboard
			}
			wantCutoff := services.ComparisonCutoff(target, comparison, test.now)

			spotlight, spotlightRepo := newController()
			recorder := dashboardLoaderRequest(spotlight.getSpotlight, target.ID.String())
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, []time.Time{test.now}, spotlightRepo.spotlightNows)

			memberships, membershipRepo := newController()
			recorder = dashboardLoaderRequest(memberships.getMembershipStats, target.ID.String())
			require.Equal(t, http.StatusOK, recorder.Code)
			if test.name == "last Toronto day across spring DST" {
				require.Equal(t, []time.Time{wantCutoff}, membershipRepo.asOfCutoffs)
			} else {
				require.Empty(t, membershipRepo.asOfCutoffs, "the exact total is used at exclusive midnight")
			}

			engagement, engagementRepo := newController()
			recorder = dashboardLoaderRequest(engagement.getEngagementStats, target.ID.String())
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, []time.Time{test.now, wantCutoff}, engagementRepo.engagementNows)

			events, eventRepo := newController()
			recorder = dashboardLoaderRequest(events.getEventActivity, target.ID.String())
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, []time.Time{test.now}, eventRepo.eventNows)
			require.Equal(t, []time.Time{wantCutoff}, eventRepo.averageNows)

			signups, signupRepo := newController()
			recorder = dashboardLoaderRequest(signups.getSignupTimeline, target.ID.String())
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, test.now, signupRepo.signupNows[0])
			require.Len(t, signupRepo.signupNows, 2, "the optional comparison loads its full dated range")
		})
	}
}

func TestSpotlightUsesTheInjectedClock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, semesters := newDashboardLoaderTestController(store.SignupTimeline{})
	dashboard := controller.store.(dashboardLoaderTestStore).dashboard
	want := time.Date(2026, time.June, 10, 18, 45, 0, 0, time.UTC)
	controller.now = func() time.Time { return want }

	recorder := dashboardLoaderRequest(controller.getSpotlight, semesters.semester.ID.String())

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []time.Time{want}, dashboard.spotlightNows)
}

func TestDashboardComparisonListFailuresAreSanitizedOrOptional(t *testing.T) {
	gin.SetMode(gin.TestMode)
	listErr := errors.New("injected semester list failure")
	tests := []struct {
		name     string
		handler  func(*dashboardController, *gin.Context)
		wantCode int
		wantBody string
		wantLog  string
	}{
		{"memberships", (*dashboardController).getMembershipStats, http.StatusInternalServerError, "unable to load dashboard data", "dashboard comparison semester lookup failed"},
		{"engagement", (*dashboardController).getEngagementStats, http.StatusInternalServerError, "unable to load dashboard data", "dashboard comparison semester lookup failed"},
		{"conversion", (*dashboardController).getTrialConversion, http.StatusInternalServerError, "unable to load dashboard data", "dashboard comparison semester lookup failed"},
		{"events", (*dashboardController).getEventActivity, http.StatusInternalServerError, "unable to load dashboard data", "dashboard comparison semester lookup failed"},
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
			require.NotContains(t, recorder.Body.String(), listErr.Error())
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
	var output bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previousOutput) })
	recorder := dashboardLoaderRequest(controller.getSignupTimeline, semesters.semester.ID.String())
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), "unable to load dashboard data")
	require.NotContains(t, recorder.Body.String(), "database password")
	require.Contains(t, output.String(), "database password and host details")
}
