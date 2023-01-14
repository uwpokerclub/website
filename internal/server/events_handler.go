package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *apiServer) ListEvents(ctx *gin.Context) {
	semesterId := ctx.Query("semesterId")

	svc := services.NewEventService(s.db)
	events, err := svc.ListEvents(semesterId)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, events)
}

func (s *apiServer) CreateEvent(ctx *gin.Context) {
	var req models.CreateEventRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewEventService(s.db)
	event, err := svc.CreateEvent(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, event)
}

func (s *apiServer) GetEvent(ctx *gin.Context) {
	eventId, err := strconv.ParseUint(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	svc := services.NewEventService(s.db)
	event, err := svc.GetEvent(eventId)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (s *apiServer) EndEvent(ctx *gin.Context) {
	eventId, err := strconv.ParseUint(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	svc := services.NewEventService(s.db)
	err = svc.EndEvent(eventId)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusNoContent, "")

}
