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
		response.Comparison = &ComparisonMembershipStats{
			Semester: store.SemesterRef{ID: comparison.ID, Name: comparison.Name},
			Stats:    comparisonStats,
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

	current, err := c.store.Dashboard().EngagementStats(semesterID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	response := EngagementStatsResponse{Current: current}

	if comparison := services.ResolveComparisonSemester(semester, semesters); comparison != nil {
		comparisonStats, err := c.store.Dashboard().EngagementStats(comparison.ID)
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
