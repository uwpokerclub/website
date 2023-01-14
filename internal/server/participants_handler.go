package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *apiServer) ListParticipants(ctx *gin.Context) {
	eventId, err := strconv.Atoi(ctx.Query("eventId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID in query"))
		return
	}

	svc := services.NewParticipantsService(s.db)
	participants, err := svc.ListParticipants(uint64(eventId))
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, participants)
}

func (s *apiServer) CreateParticipant(ctx *gin.Context) {
	var req models.CreateParticipantRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewParticipantsService(s.db)
	participant, err := svc.CreateParticipant(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, participant)
}

func (s *apiServer) SignOutParticipant(ctx *gin.Context) {
	var req models.UpdateParticipantRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}
	req.SignOut = true

	svc := services.NewParticipantsService(s.db)
	participant, err := svc.UpdateParticipant(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, participant)
}

func (s *apiServer) SignInParticipant(ctx *gin.Context) {
	var req models.UpdateParticipantRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}
	req.SignIn = true

	svc := services.NewParticipantsService(s.db)
	participant, err := svc.UpdateParticipant(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, participant)
}

func (s *apiServer) RebuyParticipant(ctx *gin.Context) {
	var req models.UpdateParticipantRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}
	req.Rebuy = true

	svc := services.NewParticipantsService(s.db)
	participant, err := svc.UpdateParticipant(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, participant)
}

func (s *apiServer) DeleteParticipant(ctx *gin.Context) {
	var req models.DeleteParticipantRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewParticipantsService(s.db)
	err = svc.DeleteParticipant(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusNoContent, "")
}
