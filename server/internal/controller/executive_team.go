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
	group.GET("", middleware.UseAuthorization("executive-team.get"), c.getExecutiveTeam)
	group.POST("", middleware.UseAuthorization("executive-team.create"), c.addExecutiveTeamMember)
	group.PATCH("/:username", middleware.UseAuthorization("executive-team.update"), c.changeExecutiveTeamRole)
	group.DELETE("/:username", middleware.UseAuthorization("executive-team.delete"), c.removeExecutiveTeamMember)
}

// removeExecutiveTeamMember removes a member from the selected semester's executive team.
//
// @Summary Remove executive team member
// @Description Disable the linked login and clear the selected semester's executive membership
// @Tags Executive Team
// @Param semesterId path string true "Semester ID"
// @Param username path string true "Login username"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/executive-team/{username} [delete]
func (c *executiveTeamController) removeExecutiveTeamMember(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest("username is required"))
		return
	}
	if err := services.NewExecutiveTeamService(c.store).RemoveMember(semesterID, username); err != nil {
		switch {
		case errors.Is(err, services.ErrExecutiveTeamProtectedLogin):
			ctx.AbortWithStatusJSON(http.StatusForbidden, apierrors.Forbidden(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamStagedLogin):
			ctx.AbortWithStatusJSON(http.StatusConflict, apierrors.Conflict(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamSemesterNotFound), errors.Is(err, store.ErrNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *executiveTeamController) changeExecutiveTeamRole(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}
	var req models.ChangeExecutiveTeamRoleRequest
	if !BindJSON(ctx, &req) {
		return
	}
	if ctx.Param("username") == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest("username is required"))
		return
	}
	if err := services.NewExecutiveTeamService(c.store).ChangeRole(semesterID, ctx.Param("username"), req.Role); err != nil {
		switch {
		case errors.Is(err, services.ErrExecutiveTeamInvalidRole), errors.Is(err, services.ErrExecutiveTeamProtectedLogin):
			ctx.AbortWithStatusJSON(http.StatusForbidden, apierrors.Forbidden(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamStagedLogin):
			ctx.AbortWithStatusJSON(http.StatusConflict, apierrors.Conflict(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamSemesterNotFound), errors.Is(err, store.ErrNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *executiveTeamController) addExecutiveTeamMember(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}
	var req models.AddExecutiveTeamMemberRequest
	if !BindJSON(ctx, &req) {
		return
	}
	response, err := services.NewExecutiveTeamService(c.store).AddMember(semesterID, req.QuestID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrExecutiveTeamInvalidIdentity):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamInvalidRole), errors.Is(err, services.ErrExecutiveTeamProtectedLogin):
			ctx.AbortWithStatusJSON(http.StatusForbidden, apierrors.Forbidden(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamMemberNotFound), errors.Is(err, services.ErrExecutiveTeamSemesterNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
		case errors.Is(err, services.ErrExecutiveTeamMemberAmbiguous), errors.Is(err, services.ErrExecutiveTeamIdentityConflict),
			errors.Is(err, services.ErrExecutiveTeamStagedLogin), errors.Is(err, services.ErrExecutiveTeamAlreadyMember),
			errors.Is(err, store.ErrConflict), errors.Is(err, store.ErrAlreadyExists):
			ctx.AbortWithStatusJSON(http.StatusConflict, apierrors.Conflict(err.Error()))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		}
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

// getExecutiveTeam returns the selected semester's executive memberships
// together with matching login details when an account exists.
//
// @Summary Get executive team roster
// @Description Return executive memberships for a semester with matching account details when available
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
	members, err := c.store.Memberships().ListExecutiveTeam(semesterID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}
	if members == nil {
		members = make([]models.ExecutiveTeamMember, 0)
	}
	ctx.JSON(http.StatusOK, models.ExecutiveTeamResponse{Members: members})
}
