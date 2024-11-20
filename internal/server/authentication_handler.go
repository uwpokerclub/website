package server

import (
	"api/internal/authentication"
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
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

func (s *apiServer) CreateLogin(ctx *gin.Context) {
	var req models.Login
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewLoginService(s.db)
	err = svc.CreateLogin(req.Username, req.Password)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

func (s *apiServer) SessionLoginHandler(ctx *gin.Context) {
	// Load and validate request body
	var req models.Login
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid body sent with request: expected {\"username\": string, \"password\": string"))
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
	cookieKey := getCookieKey()

	sessionID, _ := ctx.Cookie(cookieKey)
	sessionUUID, _ := uuid.Parse(sessionID)

	sessionManager := authentication.NewSessionManager(s.db)
	session, err := sessionManager.Get(sessionUUID)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, session)
}
