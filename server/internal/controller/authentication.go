package controller

import (
	"api/internal/authentication"
	"api/internal/authorization"
	"api/internal/middleware"
	"api/internal/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	e "api/internal/errors"
)

type authenticationController struct {
	db *gorm.DB
}

func NewAuthenticationController(db *gorm.DB) Controller {
	return &authenticationController{db: db}
}

// getCookieKey returns the key of the session ID cookie for the environment
func getCookieKey() string {
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		return "uwpsc-session-id"
	}

	return "uwpsc-dev-session-id"
}

func (controller *authenticationController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("session")
	group.GET("", middleware.UseAuthentication(controller.db), controller.getSession)
	group.POST("", controller.login)
	group.POST("logout", controller.logout)
}

// getSession retrieves the current user's session information, including username, role, and permissions.
// It requires authentication and returns a GetSessionResponse.
//
// @Summary Get current session
// @Description Retrieve current user's session information
// @Tags Authentication
// @Produce json
// @Success 200 {object} GetSessionResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /session [get]
func (controller *authenticationController) getSession(ctx *gin.Context) {
	username := ctx.GetString("username")
	role := ctx.GetString("role")

	svc := authorization.NewAuthorizationService(role, authorization.DefaultAuthorizerMap)

	ctx.JSON(http.StatusOK, models.GetSessionResponse{
		Username:    username,
		Role:        svc.Role(),
		Permissions: svc.GetPermissions(),
	})
}

// login handles user login by validating credentials and creating a session.
// It expects a NewSessionRequest in the request body and sets a session cookie upon successful login.
//
// @Summary User login
// @Description Authenticate user and create a session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body NewSessionRequest true "User credentials"
// @Success 201 "Created"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /session [post]
func (controller *authenticationController) login(ctx *gin.Context) {
	var req models.NewSessionRequest
	if !BindJSON(ctx, &req) {
		return
	}

	credentialSvc := authentication.NewCredentialService(controller.db)
	valid, role, err := credentialSvc.Validate(req.Username, req.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	if !valid {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			e.Unauthorized("Invalid username or password"),
		)
		return
	}

	sessionManager := authentication.NewSessionManager(controller.db)
	token, err := sessionManager.Create(req.Username, role)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	eightHoursInSeconds := int((time.Hour * 8).Seconds())
	ctx.SetSameSite(http.SameSiteStrictMode)
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		ctx.SetCookie(
			"uwpsc-session-id",
			token.String(),
			eightHoursInSeconds,
			"/",
			"uwpokerclub.com",
			true,
			true,
		)
	} else {
		ctx.SetCookie("uwpsc-dev-session-id", token.String(), eightHoursInSeconds, "/", "localhost", false, true)
	}

	ctx.Status(http.StatusCreated)
}

// logout handles user logout by invalidating the current session.
// It expects a valid session cookie and returns a 204 No Content status upon successful logout.
//
// @Summary User logout
// @Description Invalidate current user session
// @Tags Authentication
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /session/logout [post]
func (controller *authenticationController) logout(ctx *gin.Context) {
	cookieKey := getCookieKey()

	// Ensure cookie is in the request
	sessionID, err := ctx.Cookie(cookieKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
		return
	}

	// Ensure cookie is a valid UUID
	err = uuid.Validate(sessionID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, e.Forbidden("Invalid session ID provided"))
		return
	}

	sessionUUID, _ := uuid.Parse(sessionID)

	sessionManager := authentication.NewSessionManager(controller.db)
	err = sessionManager.Invalidate(sessionUUID)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
