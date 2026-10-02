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
	group.GET("events", middleware.UseAuthorization("semester.get"), c.getEventActivity)
	group.GET("signups", middleware.UseAuthorization("semester.get"), c.getSignupTimeline)
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
// @Description Get zero-filled daily membership creation counts by source, event dates, and the first dated signup for a semester
// @Tags Dashboard
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} SignupTimeline
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

	if _, err := c.store.Semesters().FindByID(semesterID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	timeline, err := c.store.Dashboard().SignupTimeline(semesterID, time.Now())
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, timeline)
}
