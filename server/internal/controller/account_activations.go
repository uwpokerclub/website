package controller

import (
	"api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type accountActivationsController struct{ store store.Store }

func NewAccountActivationsController(st store.Store) Controller {
	return &accountActivationsController{store: st}
}
func (c *accountActivationsController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("activations")
	group.POST("verify", middleware.RateLimit(middleware.RateLimitRequestsPerMinute()), c.verify)
	group.POST("complete", middleware.RateLimit(middleware.RateLimitRequestsPerMinute()), c.complete)
}
func (c *accountActivationsController) verify(ctx *gin.Context) {
	var req models.VerifyActivationRequest
	if !BindJSON(ctx, &req) {
		return
	}
	activation, err := services.NewAccountActivationService(c.store).Verify(req.Token)
	if err == services.ErrActivationNotFound {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("Invalid or expired activation token"))
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError(err.Error()))
		return
	}
	member, err := c.store.Logins().FindByUsernameWithMember(activation.Username)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("Invalid or expired activation token"))
		return
	}
	response := models.VerifyActivationResponse{}
	if member.LinkedMember != nil {
		response.FirstName, response.LastName = member.LinkedMember.FirstName, member.LinkedMember.LastName
	}
	ctx.JSON(http.StatusOK, response)
}
func (c *accountActivationsController) complete(ctx *gin.Context) {
	var req models.CompleteActivationRequest
	if !BindJSON(ctx, &req) {
		return
	}
	_, token, err := services.NewAccountActivationService(c.store).Complete(req.Token, req.Password)
	if err == services.ErrActivationNotFound {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("Invalid or expired activation token"))
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError(err.Error()))
		return
	}
	maxAge := int((8 * time.Hour).Seconds())
	ctx.SetSameSite(http.SameSiteStrictMode)
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		ctx.SetCookie("uwpsc-session-id", token.String(), maxAge, "/", "uwpokerclub.com", true, true)
	} else {
		ctx.SetCookie("uwpsc-dev-session-id", token.String(), maxAge, "/", "localhost", false, true)
	}
	ctx.Status(http.StatusCreated)
}
