package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
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
