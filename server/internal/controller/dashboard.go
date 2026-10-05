package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type dashboardController struct {
	store store.Store
	now   func() time.Time
}

type dashboardComparisonLoad struct {
	Semester models.Semester
	Ref      store.SemesterRef
}

// NewDashboardController creates a new instance of dashboardController.
func NewDashboardController(st store.Store) Controller {
	return &dashboardController{store: st, now: time.Now}
}

func (c *dashboardController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters/:semesterId/dashboard", middleware.UseAuthentication(c.store))
	group.GET("spotlight", middleware.UseAuthorization("semester.get"), c.getSpotlight)
	group.GET("memberships", middleware.UseAuthorization("semester.get"), c.getMembershipStats)
	group.GET("engagement", middleware.UseAuthorization("semester.get"), c.getEngagementStats)
	group.GET("conversion", middleware.UseAuthorization("semester.get"), c.getTrialConversion)
	group.GET("events", middleware.UseAuthorization("semester.get"), c.getEventActivity)
	group.GET("signups", middleware.UseAuthorization("semester.get"), c.getSignupTimeline)
}

func (c *dashboardController) loadDashboardSemester(ctx *gin.Context) (models.Semester, bool) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return models.Semester{}, false
	}

	semester, err := c.store.Semesters().FindByID(semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return models.Semester{}, false
		}
		abortDashboardInternalError(ctx, "semester lookup", err)
		return models.Semester{}, false
	}

	return semester, true
}

func (c *dashboardController) resolveDashboardComparison(semester models.Semester) (*dashboardComparisonLoad, error) {
	semesters, _, err := c.store.Semesters().List(&models.Pagination{})
	if err != nil {
		return nil, err
	}
	comparison := services.ResolveComparisonSemester(semester, semesters)
	if comparison == nil {
		return nil, nil
	}
	return &dashboardComparisonLoad{
		Semester: *comparison,
		Ref:      store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
	}, nil
}

// loadDashboardComparison is for routes where a comparison is required for the
// response. Signup comparison uses the resolver directly because that overlay is
// optional and must not abort an otherwise successful timeline response.
func (c *dashboardController) loadDashboardComparison(
	ctx *gin.Context,
	semester models.Semester,
) (*dashboardComparisonLoad, bool) {
	comparison, err := c.resolveDashboardComparison(semester)
	if err != nil {
		abortDashboardInternalError(ctx, "comparison semester lookup", err)
		return nil, false
	}
	return comparison, true
}

func abortDashboardInternalError(ctx *gin.Context, operation string, err error) {
	log.Printf("dashboard %s failed (semester_id=%s): %v", operation, ctx.Param("semesterId"), err)
	ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError("unable to load dashboard data"))
}

// ComparisonTrialConversionStats pairs a resolved comparison semester with its
// trial-conversion stats, calculated using that semester's own free-trial limit.
type ComparisonTrialConversionStats struct {
	Semester       store.SemesterRef                `json:"semester"`
	Stats          store.TrialConversionStats       `json:"stats"`
	Conversion     store.TrialConversionCohortStats `json:"conversion"`
	FreeTrialLimit uint8                            `json:"freeTrialLimit"`
} //@name ComparisonTrialConversionStats

// TrialConversionResponse is the dashboard's Trial Conversion response.
type TrialConversionResponse struct {
	Current        store.TrialConversionStats       `json:"current"`
	Conversion     store.TrialConversionCohortStats `json:"conversion"`
	FreeTrialLimit uint8                            `json:"freeTrialLimit"`
	Comparison     *ComparisonTrialConversionStats  `json:"comparison"`
} //@name TrialConversionResponse

// getTrialConversion handles retrieving the dashboard's Trial Conversion card for a semester.
//
// @Summary Get dashboard trial conversion stats
// @Description Get the distinct-player paid, executive, and free-trial status breakdown for a semester, plus the same figures for the resolved comparison semester, or null when there is no comparable term
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} TrialConversionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/dashboard/conversion [get]
func (c *dashboardController) getTrialConversion(ctx *gin.Context) {
	semester, ok := c.loadDashboardSemester(ctx)
	if !ok {
		return
	}
	semesterID := semester.ID
	comparison, ok := c.loadDashboardComparison(ctx, semester)
	if !ok {
		return
	}

	stats, err := c.store.Dashboard().TrialConversionStats(semesterID, semester.FreeTrialLimit)
	if err != nil {
		abortDashboardInternalError(ctx, "trial conversion query", err)
		return
	}

	conversion, err := c.store.Dashboard().TrialConversionCohortStats(semesterID)
	if err != nil {
		abortDashboardInternalError(ctx, "trial conversion cohort query", err)
		return
	}
	response := TrialConversionResponse{Current: stats, Conversion: conversion, FreeTrialLimit: semester.FreeTrialLimit}

	if comparison != nil {
		comparisonStats, err := c.store.Dashboard().TrialConversionStats(comparison.Semester.ID, comparison.Semester.FreeTrialLimit)
		if err != nil {
			abortDashboardInternalError(ctx, "comparison trial conversion query", err)
			return
		}
		comparisonConversion, err := c.store.Dashboard().TrialConversionCohortStats(comparison.Semester.ID)
		if err != nil {
			abortDashboardInternalError(ctx, "comparison trial conversion cohort query", err)
			return
		}
		response.Comparison = &ComparisonTrialConversionStats{
			Semester:       comparison.Ref,
			Stats:          comparisonStats,
			Conversion:     comparisonConversion,
			FreeTrialLimit: comparison.Semester.FreeTrialLimit,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// getSpotlight handles retrieving the dashboard's Event Spotlight card for a semester.
//
// @Summary Get dashboard spotlight
// @Description Get the live or next upcoming event for a semester, or null if neither exists
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} SpotlightEvent "The live or next upcoming event, or null if neither exists"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/dashboard/spotlight [get]
func (c *dashboardController) getSpotlight(ctx *gin.Context) {
	semester, ok := c.loadDashboardSemester(ctx)
	if !ok {
		return
	}
	semesterID := semester.ID

	now := c.now()
	event, err := c.store.Dashboard().Spotlight(semesterID, now)
	if err != nil {
		abortDashboardInternalError(ctx, "spotlight query", err)
		return
	}

	ctx.JSON(http.StatusOK, event)
}

// ComparisonMembershipStats pairs a resolved comparison semester with its membership
// stats.
type ComparisonMembershipStats struct {
	Semester store.SemesterRef     `json:"semester"`
	Stats    store.MembershipStats `json:"stats"`

	// TotalAsOf is the comparison term's membership count at the same elapsed point
	// this term has reached, or its exact final total when the comparison period is
	// complete. A partial count is the exact observed dated count, can omit undated
	// memberships, and is never scaled. It is null when the all-term dated share is
	// below the repository's reliability threshold. Null means unknowable, never zero.
	TotalAsOf *int64 `json:"totalAsOf"`
} //@name ComparisonMembershipStats

// MembershipStatsResponse is the dashboard's Term at a Glance response: the current
// semester's membership stats, plus the same figures for the resolved comparison
// semester, or null when there is no comparable term.
type MembershipStatsResponse struct {
	Current    store.MembershipStats      `json:"current"`
	Comparison *ComparisonMembershipStats `json:"comparison"`
} //@name MembershipStatsResponse

// getMembershipStats handles retrieving the dashboard's Term at a Glance card for a
// semester.
//
// @Summary Get dashboard membership stats
// @Description Get membership counts by paid/unpaid/discounted/executive and the new-vs-returning split for a semester, plus the same figures for the resolved comparison semester, or null when there is no comparable term
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} MembershipStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/dashboard/memberships [get]
func (c *dashboardController) getMembershipStats(ctx *gin.Context) {
	semester, ok := c.loadDashboardSemester(ctx)
	if !ok {
		return
	}
	semesterID := semester.ID

	comparison, ok := c.loadDashboardComparison(ctx, semester)
	if !ok {
		return
	}
	now := c.now()

	current, err := c.store.Dashboard().MembershipStats(semesterID)
	if err != nil {
		abortDashboardInternalError(ctx, "membership stats query", err)
		return
	}

	response := MembershipStatsResponse{Current: current}

	if comparison != nil {
		comparisonStats, err := c.store.Dashboard().MembershipStats(comparison.Semester.ID)
		if err != nil {
			abortDashboardInternalError(ctx, "comparison membership stats query", err)
			return
		}

		// Memberships cannot be clipped the way events can: rows predating the
		// created_at migration carry NULL and their dates are unrecoverable. For a
		// partial comparison, only return a dated count when the store's reliability
		// threshold is met. Once the comparison period is complete, its exact final
		// total is known from MembershipStats and includes undated rows.
		cutoff := services.ComparisonCutoff(semester, comparison.Semester, now)
		var totalAsOf *int64
		if services.IsComparisonComplete(semester, comparison.Semester, now) {
			total := comparisonStats.Total
			totalAsOf = &total
		} else {
			totalAsOf, err = c.store.Dashboard().MembershipTotalAsOf(comparison.Semester.ID, cutoff)
			if err != nil {
				abortDashboardInternalError(ctx, "comparison membership cutoff query", err)
				return
			}
		}

		response.Comparison = &ComparisonMembershipStats{
			Semester:  comparison.Ref,
			Stats:     comparisonStats,
			TotalAsOf: totalAsOf,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// ComparisonEngagementStats pairs a resolved comparison semester with its engagement
// stats.
type ComparisonEngagementStats struct {
	Semester store.SemesterRef     `json:"semester"`
	Stats    store.EngagementStats `json:"stats"`
} //@name ComparisonEngagementStats

// EngagementStatsResponse is the dashboard's Engagement & Retention response: the
// current semester's engagement stats, plus the same figures for the resolved
// comparison semester, or null when there is no comparable term.
type EngagementStatsResponse struct {
	Current    store.EngagementStats      `json:"current"`
	Comparison *ComparisonEngagementStats `json:"comparison"`
} //@name EngagementStatsResponse

// getEngagementStats handles retrieving the dashboard's Engagement & Retention card
// for a semester.
//
// @Summary Get dashboard engagement stats
// @Description Get distinct players, median events attended, played-once share, and the 10+ cohort size for a semester, plus the same figures for the resolved comparison semester, or null when there is no comparable term
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} EngagementStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/dashboard/engagement [get]
func (c *dashboardController) getEngagementStats(ctx *gin.Context) {
	semester, ok := c.loadDashboardSemester(ctx)
	if !ok {
		return
	}
	semesterID := semester.ID

	comparison, ok := c.loadDashboardComparison(ctx, semester)
	if !ok {
		return
	}

	now := c.now()

	current, err := c.store.Dashboard().EngagementStats(semesterID, now)
	if err != nil {
		abortDashboardInternalError(ctx, "engagement stats query", err)
		return
	}

	response := EngagementStatsResponse{Current: current}

	if comparison != nil {
		// Clipped to the same elapsed point, so a term a month in is not measured
		// against four completed months of the previous one.
		cutoff := services.ComparisonCutoff(semester, comparison.Semester, now)
		comparisonStats, err := c.store.Dashboard().EngagementStats(comparison.Semester.ID, cutoff)
		if err != nil {
			abortDashboardInternalError(ctx, "comparison engagement stats query", err)
			return
		}
		response.Comparison = &ComparisonEngagementStats{
			Semester: comparison.Ref,
			Stats:    comparisonStats,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// EventActivityCurrent is the current semester's Event Activity figures and series.
type EventActivityCurrent struct {
	store.EventActivityStats
	Series []store.EventSeriesPoint `json:"series"`
} //@name EventActivityCurrent

// ComparisonEventActivity is the comparison semester and its average field size.
// AverageFieldSize is null when no events had ended by the comparison cutoff; zero
// means completed events existed but had no entries.
type ComparisonEventActivity struct {
	Semester store.SemesterRef `json:"semester"`
	// AverageFieldSize is null when no events had ended by the comparison cutoff;
	// zero means completed events existed but had no entries.
	AverageFieldSize *float64 `json:"averageFieldSize" extensions:"x-nullable"`
} //@name ComparisonEventActivity

// EventActivityResponse is the dashboard's Event Activity response.
type EventActivityResponse struct {
	Current    EventActivityCurrent     `json:"current"`
	Comparison *ComparisonEventActivity `json:"comparison"`
} //@name EventActivityResponse

// SignupTimelineComparisonPoint is the comparison semester's daily signup total
// at an elapsed calendar-day offset from its own start date.
type SignupTimelineComparisonPoint struct {
	ElapsedDay int   `json:"elapsedDay"`
	Total      int64 `json:"total"`
} //@name SignupTimelineComparisonPoint

// SignupTimelineComparison contains the resolved previous same-season semester
// and its dated daily signup totals. Missing offsets are outside term coverage.
type SignupTimelineComparison struct {
	Semester    store.SemesterRef               `json:"semester"`
	DailyTotals []SignupTimelineComparisonPoint `json:"dailyTotals"`
} //@name SignupTimelineComparison

// SignupTimelineResponse is the current signup timeline plus an optional prior
// same-season daily-total trend.
type SignupTimelineResponse struct {
	store.SignupTimeline
	Comparison *SignupTimelineComparison `json:"comparison"`
} //@name SignupTimelineResponse

// getEventActivity handles retrieving the dashboard's Event Activity card for a semester.
//
// @Summary Get dashboard event activity
// @Description Get events run vs scheduled, total entries, average field size, and per-event attendance for a semester, plus the comparison semester and its average field size, or null when there is no comparable term
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} EventActivityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/dashboard/events [get]
func (c *dashboardController) getEventActivity(ctx *gin.Context) {
	semester, ok := c.loadDashboardSemester(ctx)
	if !ok {
		return
	}
	semesterID := semester.ID

	comparison, ok := c.loadDashboardComparison(ctx, semester)
	if !ok {
		return
	}

	now := c.now()

	stats, series, err := c.store.Dashboard().EventActivity(semesterID, now)
	if err != nil {
		abortDashboardInternalError(ctx, "event activity query", err)
		return
	}

	response := EventActivityResponse{Current: EventActivityCurrent{EventActivityStats: stats, Series: series}}
	if comparison != nil {
		cutoff := services.ComparisonCutoff(semester, comparison.Semester, now)
		comparisonAverage, err := c.store.Dashboard().AverageFieldSize(comparison.Semester.ID, cutoff)
		if err != nil {
			abortDashboardInternalError(ctx, "comparison event average query", err)
			return
		}
		response.Comparison = &ComparisonEventActivity{
			Semester:         comparison.Ref,
			AverageFieldSize: comparisonAverage,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// getSignupTimeline handles retrieving daily membership creation counts for a semester.
//
// @Summary Get dashboard signup timeline
// @Description Get zero-filled daily membership creation counts by source and event dates, plus an optional prior same-season daily-total trend aligned by elapsed calendar day
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} SignupTimelineResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/dashboard/signups [get]
func (c *dashboardController) getSignupTimeline(ctx *gin.Context) {
	semester, ok := c.loadDashboardSemester(ctx)
	if !ok {
		return
	}
	semesterID := semester.ID

	now := c.now()
	timeline, err := c.store.Dashboard().SignupTimeline(semesterID, now)
	if err != nil {
		if errors.Is(err, store.ErrSignupTimelineRange) {
			abortDashboardInternalError(ctx, "signup timeline range", err)
			return
		}
		abortDashboardInternalError(ctx, "signup timeline query", err)
		return
	}

	response := SignupTimelineResponse{SignupTimeline: timeline}
	if len(timeline.Series) > 0 {
		comparison, err := c.resolveDashboardComparison(semester)
		if err != nil {
			log.Printf("dashboard signup comparison semester lookup failed (semester_id=%s): %v", semesterID, err)
		} else if comparison != nil {
			// The repository's date-only series is the source of truth for both
			// semester bounds; do not reinterpret a start timestamp in Toronto.
			comparisonAsOf := time.Date(9999, 12, 31, 12, 0, 0, 0, time.UTC)
			fullComparison, err := c.store.Dashboard().SignupTimeline(comparison.Semester.ID, comparisonAsOf)
			if err != nil {
				if errors.Is(err, store.ErrSignupTimelineRange) {
					log.Printf("dashboard signup comparison overlay unavailable: comparison range exceeds limit (semester_id=%s comparison_semester_id=%s)", semesterID, comparison.Semester.ID)
				} else {
					log.Printf("dashboard signup comparison timeline query failed (semester_id=%s comparison_semester_id=%s): %v", semesterID, comparison.Semester.ID, err)
				}
			} else if fullComparison.DataStartsAt != nil && len(fullComparison.Series) > 0 {
				// Reuse MembershipTotalAsOf's all-term dated-share gate. The visible
				// elapsed series cannot represent the denominator: it omits later and
				// otherwise undated memberships.
				datedTotal, err := c.store.Dashboard().MembershipTotalAsOf(comparison.Semester.ID, comparisonAsOf)
				if err != nil {
					log.Printf("dashboard signup comparison coverage query failed (semester_id=%s comparison_semester_id=%s): %v", semesterID, comparison.Semester.ID, err)
				} else if datedTotal != nil {
					points, ok := signupComparisonDailyTotals(timeline.Series, fullComparison.Series)
					if ok {
						response.Comparison = &SignupTimelineComparison{
							Semester:    comparison.Ref,
							DailyTotals: points,
						}
					}
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// signupComparisonDailyTotals aligns the full prior-term series to the visible
// current calendar-day span. Both starts come from the repository's YYYY-MM-DD
// rows, so database-session date casts remain authoritative; parsing those
// date-only values in UTC makes subtraction immune to DST day lengths.
func signupComparisonDailyTotals(
	currentSeries, comparisonSeries []store.SignupTimelinePoint,
) ([]SignupTimelineComparisonPoint, bool) {
	if len(currentSeries) == 0 || len(comparisonSeries) == 0 {
		return nil, false
	}
	currentStart, err := time.Parse("2006-01-02", currentSeries[0].Date)
	if err != nil {
		return nil, false
	}
	currentEnd, err := time.Parse("2006-01-02", currentSeries[len(currentSeries)-1].Date)
	if err != nil {
		return nil, false
	}
	currentSpan := int(currentEnd.Sub(currentStart).Hours() / 24)
	if currentSpan < 0 {
		return nil, false
	}

	comparisonStart, err := time.Parse("2006-01-02", comparisonSeries[0].Date)
	if err != nil {
		return nil, false
	}
	points := make([]SignupTimelineComparisonPoint, 0, len(comparisonSeries))
	for _, point := range comparisonSeries {
		date, err := time.Parse("2006-01-02", point.Date)
		if err != nil {
			return nil, false
		}
		elapsedDay := int(date.Sub(comparisonStart).Hours() / 24)
		if elapsedDay > currentSpan {
			break
		}
		if elapsedDay < 0 {
			continue
		}
		points = append(points, SignupTimelineComparisonPoint{
			ElapsedDay: elapsedDay,
			Total:      point.Admin + point.Discord + point.Unknown,
		})
	}
	return points, true
}
