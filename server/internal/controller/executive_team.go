package controller

import (
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"net/http"

	apierrors "api/internal/errors"
	"github.com/gin-gonic/gin"
)

type executiveTeamController struct{ store store.Store }

func NewExecutiveTeamController(st store.Store) Controller {
	return &executiveTeamController{store: st}
}

func (c *executiveTeamController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters/:semesterId/executive-team", middleware.UseAuthentication(c.store))
	group.GET("", middleware.UseAuthorization("executive_team.get"), c.getExecutiveTeam)
}

// getExecutiveTeam returns the selected semester's executive memberships
// together with currently eligible executive login accounts.
//
// @Summary Get executive team roster
// @Description Return executive memberships for a semester unioned with current eligible executive logins
// @Tags Executive Team
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {object} ExecutiveTeamResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/executive-team [get]
func (c *executiveTeamController) getExecutiveTeam(ctx *gin.Context) {
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
	response, err := services.GetExecutiveTeam(c.store, semesterID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}
	if response.Members == nil {
		response.Members = make([]models.ExecutiveTeamMember, 0)
	}
	ctx.JSON(http.StatusOK, response)
}
