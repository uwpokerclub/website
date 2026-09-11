package controller

import (
	"api/internal/authorization"
	"api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	stderrors "errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	g.POST(":id/cancel", middleware.UseAuthorization("officer-transition.cancel"), c.cancel)
	g.POST(":id/reissue", middleware.UseAuthorization("officer-transition.reissue"), c.reissue)
}

func transitionID(ctx *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.InvalidRequest("invalid officer transition ID"))
		return uuid.Nil, false
	}
	return id, true
}

func (c *officerTransitionsController) cancel(ctx *gin.Context) {
	id, ok := transitionID(ctx)
	if !ok {
		return
	}
	_, err := services.NewOfficerTransitionService(c.store).Cancel(id)
	if err != nil {
		switch {
		case stderrors.Is(err, services.ErrTransitionNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, errors.NotFound(err.Error()))
		case stderrors.Is(err, services.ErrTransitionResolved):
			ctx.AbortWithStatusJSON(http.StatusConflict, errors.Conflict(err.Error()))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError(err.Error()))
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *officerTransitionsController) reissue(ctx *gin.Context) {
	id, ok := transitionID(ctx)
	if !ok {
		return
	}
	var req models.ReissueOfficerTransitionRequest
	if !BindJSON(ctx, &req) {
		return
	}
	token, err := services.NewOfficerTransitionService(c.store).Reissue(id, authorization.ToRole(req.Role))
	if err != nil {
		switch {
		case stderrors.Is(err, services.ErrTransitionInvalid):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.InvalidRequest(err.Error()))
		case stderrors.Is(err, services.ErrTransitionNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, errors.NotFound(err.Error()))
		case stderrors.Is(err, services.ErrTransitionResolved), stderrors.Is(err, services.ErrTransitionIneligible):
			ctx.AbortWithStatusJSON(http.StatusConflict, errors.Conflict(err.Error()))
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError(err.Error()))
		}
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"activationToken": token})
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
