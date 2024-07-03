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
)

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

// TODO: remove this once session-based authentication implemented in app and server
func (s *apiServer) NewSession(ctx *gin.Context) {
	var req models.Login
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewLoginService(s.db)
	token, err := svc.ValidateCredentials(req.Username, req.Password)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	// Set cookie
	oneDayInSeconds := 86400

	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		ctx.SetCookie("pctoken", token, oneDayInSeconds, "/", "uwpokerclub.com", true, false)
	} else {
		ctx.SetCookie("pctoken", token, oneDayInSeconds, "/", "localhost", false, false)
	}

	// Return empty created response
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
