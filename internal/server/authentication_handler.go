package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"os"
	"strings"

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
