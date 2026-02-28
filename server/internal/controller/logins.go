package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type loginsController struct {
	db *gorm.DB
}

// NewLoginsController creates a new instance of loginsController
func NewLoginsController(db *gorm.DB) Controller {
	return &loginsController{db: db}
}

func (c *loginsController) LoadRoutes(router *gin.RouterGroup) {
	logins := router.Group("logins", middleware.UseAuthentication(c.db))
	logins.GET("", middleware.UseAuthorization("login.list"), c.listLogins)
	logins.GET("/:username", middleware.UseAuthorization("login.get"), c.getLogin)
	logins.POST("", middleware.UseAuthorization("login.create"), c.createLogin)
	logins.DELETE("/:username", middleware.UseAuthorization("login.delete"), c.deleteLogin)
	logins.PATCH("/:username/password", middleware.UseAuthorization("login.edit"), c.changePassword)
}

// listLogins handles listing all logins with linked member information
//
// @Summary List all logins
// @Description Retrieve a list of all logins with their linked member information
// @Tags Logins
// @Accept json
// @Produce json
// @Success 200 {array} LoginWithMember
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /logins [get]
func (c *loginsController) listLogins(ctx *gin.Context) {
	pagination, err := models.ParsePagination(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewLoginService(c.db)
	logins, total, err := svc.ListLogins(&pagination)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, models.ListResponse[models.LoginWithMember]{
		Data:  logins,
		Total: total,
	})
}

// getLogin handles retrieving a single login by username
//
// @Summary Get login by username
// @Description Retrieve a login by username with linked member information
// @Tags Logins
// @Accept json
// @Produce json
// @Param username path string true "Login username"
// @Success 200 {object} LoginWithMember
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /logins/{username} [get]
func (c *loginsController) getLogin(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest("username parameter is required"),
		)
		return
	}

	svc := services.NewLoginService(c.db)
	login, err := svc.GetLogin(username)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, login)
}

// createLogin handles the creation of a new login
//
// @Summary Create a new login
// @Description Create a new login with the provided credentials
// @Tags Logins
// @Accept json
// @Produce json
// @Param login body CreateLoginRequest true "Login details"
// @Success 201 {object} LoginWithMember
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /logins [post]
func (c *loginsController) createLogin(ctx *gin.Context) {
	var req models.CreateLoginRequest
	if !BindJSON(ctx, &req) {
		return
	}

	svc := services.NewLoginService(c.db)
	err := svc.CreateLoginFromRequest(&req)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	// Return the created login (without password)
	login, err := svc.GetLogin(req.Username)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusCreated, login)
}

// deleteLogin handles deleting a login by username
//
// @Summary Delete login by username
// @Description Delete a login by username
// @Tags Logins
// @Accept json
// @Produce json
// @Param username path string true "Login username"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /logins/{username} [delete]
func (c *loginsController) deleteLogin(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest("username parameter is required"),
		)
		return
	}

	svc := services.NewLoginService(c.db)
	err := svc.DeleteLogin(username)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// changePassword handles changing a login's password
//
// @Summary Change login password
// @Description Change the password for a login
// @Tags Logins
// @Accept json
// @Produce json
// @Param username path string true "Login username"
// @Param request body ChangePasswordRequest true "New password"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /logins/{username}/password [patch]
func (c *loginsController) changePassword(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest("username parameter is required"),
		)
		return
	}

	var req models.ChangePasswordRequest
	if !BindJSON(ctx, &req) {
		return
	}

	svc := services.NewLoginService(c.db)
	err := svc.ChangePassword(username, req.NewPassword)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}
