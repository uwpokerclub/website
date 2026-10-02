package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type dashboardController struct {
	store store.Store
}

// NewDashboardController creates a new instance of dashboardController.
func NewDashboardController(st store.Store) Controller {
	return &dashboardController{store: st}
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
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	semester, err := c.store.Semesters().FindByID(semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	stats, err := c.store.Dashboard().TrialConversionStats(semesterID, semester.FreeTrialLimit)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	conversion, err := c.store.Dashboard().TrialConversionCohortStats(semesterID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}
	response := TrialConversionResponse{Current: stats, Conversion: conversion, FreeTrialLimit: semester.FreeTrialLimit}

	semesters, _, err := c.store.Semesters().List(&models.Pagination{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	if comparison := services.ResolveComparisonSemester(semester, semesters); comparison != nil {
		comparisonStats, err := c.store.Dashboard().TrialConversionStats(comparison.ID, comparison.FreeTrialLimit)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}
		comparisonConversion, err := c.store.Dashboard().TrialConversionCohortStats(comparison.ID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}
		response.Comparison = &ComparisonTrialConversionStats{
			Semester:       store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
			Stats:          comparisonStats,
			Conversion:     comparisonConversion,
			FreeTrialLimit: comparison.FreeTrialLimit,
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
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	if _, err := c.store.Semesters().FindByID(semesterID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	event, err := c.store.Dashboard().Spotlight(semesterID, time.Now())
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
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
	// this term has reached, or null when that term has no dated memberships. Null
	// means unknowable, never zero.
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
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	semester, err := c.store.Semesters().FindByID(semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	semesters, _, err := c.store.Semesters().List(&models.Pagination{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	current, err := c.store.Dashboard().MembershipStats(semesterID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	response := MembershipStatsResponse{Current: current}

	if comparison := services.ResolveComparisonSemester(semester, semesters); comparison != nil {
		comparisonStats, err := c.store.Dashboard().MembershipStats(comparison.ID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}

		// Memberships cannot be clipped the way events can: rows predating the
		// created_at migration carry NULL and their dates are unrecoverable. TotalAsOf
		// is nil for those terms, and the UI shows pace against the final total rather
		// than a delta against a figure it cannot compute.
		cutoff := services.ComparisonCutoff(semester, *comparison, time.Now())
		totalAsOf, err := c.store.Dashboard().MembershipTotalAsOf(comparison.ID, cutoff)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}

		response.Comparison = &ComparisonMembershipStats{
			Semester:  store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
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
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	semester, err := c.store.Semesters().FindByID(semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	semesters, _, err := c.store.Semesters().List(&models.Pagination{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	now := time.Now()

	current, err := c.store.Dashboard().EngagementStats(semesterID, now)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	response := EngagementStatsResponse{Current: current}

	if comparison := services.ResolveComparisonSemester(semester, semesters); comparison != nil {
		// Clipped to the same elapsed point, so a term a month in is not measured
		// against four completed months of the previous one.
		cutoff := services.ComparisonCutoff(semester, *comparison, now)
		comparisonStats, err := c.store.Dashboard().EngagementStats(comparison.ID, cutoff)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}
		response.Comparison = &ComparisonEngagementStats{
			Semester: store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
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
type ComparisonEventActivity struct {
	Semester         store.SemesterRef `json:"semester"`
	AverageFieldSize float64           `json:"averageFieldSize"`
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
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	semester, err := c.store.Semesters().FindByID(semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	semesters, _, err := c.store.Semesters().List(&models.Pagination{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	now := time.Now()

	stats, series, err := c.store.Dashboard().EventActivity(semesterID, now)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	response := EventActivityResponse{Current: EventActivityCurrent{EventActivityStats: stats, Series: series}}
	if comparison := services.ResolveComparisonSemester(semester, semesters); comparison != nil {
		cutoff := services.ComparisonCutoff(semester, *comparison, now)
		comparisonStats, _, err := c.store.Dashboard().EventActivity(comparison.ID, cutoff)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}
		response.Comparison = &ComparisonEventActivity{
			Semester:         store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
			AverageFieldSize: comparisonStats.AverageFieldSize,
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
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	semester, err := c.store.Semesters().FindByID(semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	now := time.Now()
	timeline, err := c.store.Dashboard().SignupTimeline(semesterID, now)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	response := SignupTimelineResponse{SignupTimeline: timeline}
	if len(timeline.Series) > 0 {
		semesters, _, err := c.store.Semesters().List(&models.Pagination{})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
			return
		}
		if comparison := services.ResolveComparisonSemester(semester, semesters); comparison != nil {
			comparisonNow, ok := signupComparisonCutoff(semester, *comparison, timeline.Series)
			if ok {
				// A full-term read distinguishes a genuinely dated zero-signup span
				// from legacy memberships whose creation dates are unknowable.
				fullComparison, err := c.store.Dashboard().SignupTimeline(comparison.ID, time.Date(9999, 12, 31, 12, 0, 0, 0, time.UTC))
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
					return
				}
				if fullComparison.DataStartsAt != nil {
					comparisonTimeline, err := c.store.Dashboard().SignupTimeline(comparison.ID, comparisonNow)
					if err != nil {
						ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
						return
					}
					if len(comparisonTimeline.Series) == 0 {
						response.Comparison = &SignupTimelineComparison{
							Semester:    store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
							DailyTotals: []SignupTimelineComparisonPoint{},
						}
					} else {
						startDate, err := time.Parse("2006-01-02", comparisonTimeline.Series[0].Date)
						if err != nil {
							ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
							return
						}
						points := make([]SignupTimelineComparisonPoint, 0, len(comparisonTimeline.Series))
						for _, point := range comparisonTimeline.Series {
							date, err := time.Parse("2006-01-02", point.Date)
							if err != nil {
								ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
								return
							}
							points = append(points, SignupTimelineComparisonPoint{
								ElapsedDay: int(date.Sub(startDate).Hours() / 24),
								Total:      point.Admin + point.Discord + point.Unknown,
							})
						}
						response.Comparison = &SignupTimelineComparison{
							Semester:    store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
							DailyTotals: points,
						}
					}
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// signupComparisonCutoff converts the last displayed current calendar date into
// the corresponding comparison calendar date. ComparisonCutoff supplies the
// existing elapsed-time bound; date-only UTC arithmetic then avoids 23/25-hour
// DST days and noon in Toronto ensures the repository's Toronto-date cast keeps
// the selected date unchanged.
func signupComparisonCutoff(current, comparison models.Semester, currentSeries []store.SignupTimelinePoint) (time.Time, bool) {
	if len(currentSeries) == 0 {
		return time.Time{}, false
	}
	currentStart, err := time.Parse("2006-01-02", currentSeries[0].Date)
	if err != nil {
		return time.Time{}, false
	}
	lastDate, err := time.Parse("2006-01-02", currentSeries[len(currentSeries)-1].Date)
	if err != nil {
		return time.Time{}, false
	}
	dayOffset := int(lastDate.Sub(currentStart).Hours() / 24)
	if dayOffset < 0 {
		return time.Time{}, false
	}

	// Preserve the target start instant's time component so ComparisonCutoff
	// receives an exact whole-calendar-day span, irrespective of DST.
	selectedCurrentDay := current.StartDate.UTC().AddDate(0, 0, dayOffset)
	cutoff := services.ComparisonCutoff(current, comparison, selectedCurrentDay)
	comparisonDate := comparison.StartDate.UTC().AddDate(0, 0, dayOffset)
	if cutoff.Year() == 9999 {
		comparisonDate = comparison.EndDate.UTC()
	}
	toronto, err := time.LoadLocation("America/Toronto")
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(comparisonDate.Year(), comparisonDate.Month(), comparisonDate.Day(), 12, 0, 0, 0, toronto), true
}
