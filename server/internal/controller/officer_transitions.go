package controller

import (
	"api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	stderrors "errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type officerTransitionsController struct{ store store.Store }

func NewOfficerTransitionsController(s store.Store) Controller {
	return &officerTransitionsController{s}
}
func (c *officerTransitionsController) LoadRoutes(r *gin.RouterGroup) {
	g := r.Group("officer-transitions", middleware.UseAuthentication(c.store))
	g.POST("", middleware.UseAuthorization("officer-transition.create"), c.create)
	g.GET("current", middleware.UseAuthorization("officer-transition.get"), c.current)
}
func (c *officerTransitionsController) create(ctx *gin.Context) {
	var req models.CreateOfficerTransitionRequest
	if !BindJSON(ctx, &req) {
		return
	}
	transition, tokens, err := services.NewOfficerTransitionService(c.store).Create(ctx.GetString("username"), req)
	if err != nil {
		switch {
		case stderrors.Is(err, services.ErrTransitionInvalid):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.InvalidRequest(err.Error()))
		case stderrors.Is(err, services.ErrTransitionForbidden):
			ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Forbidden(err.Error()))
		case stderrors.Is(err, services.ErrTransitionPending):
			ctx.AbortWithStatusJSON(http.StatusConflict, errors.Conflict(err.Error()))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError(err.Error()))
		}
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"transition": transition, "activationTokens": tokens})
}
func (c *officerTransitionsController) current(ctx *gin.Context) {
	transition, err := c.store.OfficerTransitions().Current()
	if stderrors.Is(err, store.ErrNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.NotFound("no pending officer transition"))
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, transition)
}
