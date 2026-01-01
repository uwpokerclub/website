package server

import (
	"api/internal/authentication"
	"api/internal/authorization"
	e "api/internal/errors"
	"api/internal/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// / getCookieKey returns the key of the session ID cookie for the environment
func getCookieKey() string {
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		return "uwpsc-session-id"
	}

	return "uwpsc-dev-session-id"
}

func (s *apiServer) SessionLoginHandler(ctx *gin.Context) {
	// Load and validate request body
	var req models.NewSessionRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	// Valdiate that the credentials provided are valid credentials
	credentialSvc := authentication.NewCredentialService(s.db)
	valid, err := credentialSvc.Validate(req.Username, req.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	if !valid {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Invalid username/password provided"))
		return
	}

	// Once credentials have been validated, create a new session in the database
	sessionManager := authentication.NewSessionManager(s.db)
	token, err := sessionManager.Create(req.Username)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	// Set cookie in response
	eightHoursInSeconds := int((time.Hour * 8).Seconds())
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		ctx.SetSameSite(http.SameSiteStrictMode)
		ctx.SetCookie("uwpsc-session-id", token.String(), eightHoursInSeconds, "/", "uwpokerclub.com", true, true)
	} else {
		ctx.SetSameSite(http.SameSiteStrictMode)
		ctx.SetCookie("uwpsc-dev-session-id", token.String(), eightHoursInSeconds, "/", "localhost", false, true)
	}

	// Returm an empty response with status code 201
	ctx.Status(http.StatusCreated)
}

func (s *apiServer) SessionLogoutHandler(ctx *gin.Context) {
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

	sessionManger := authentication.NewSessionManager(s.db)
	err = sessionManger.Invalidate(sessionUUID)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *apiServer) GetSessionHandler(ctx *gin.Context) {
	username := ctx.GetString("username")

	svc, err := authorization.NewAuthorizationService(s.db, username, authorization.DefaultAuthorizerMap)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, models.GetSessionResponse{
		Username:    username,
		Role:        svc.Role(),
		Permissions: svc.GetPermissions(),
	})
}
