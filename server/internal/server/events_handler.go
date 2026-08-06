package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *apiServer) ListEvents(ctx *gin.Context) {
	semesterIdParam := ctx.Query("semesterId")

	filter := &models.ListEventsFilter{Pagination: models.Pagination{}}
	if semesterIdParam != "" {
		semesterUUID, err := uuid.Parse(semesterIdParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid semester ID specified in request"))
			return
		}
		filter.SemesterID = &semesterUUID
	}

	events, _, err := s.store.Events().List(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ret := make([]models.ListEventsResponse, len(events))
	for i, ev := range events {
		ret[i] = models.ListEventsResponse{
			ID:         ev.ID,
			Name:       ev.Name,
			Format:     ev.Format,
			Notes:      ev.Notes,
			SemesterID: ev.SemesterID.String(),
			StartDate:  ev.StartDate,
			State:      ev.State,
			Count:      int32(len(ev.Entries)),
		}
	}

	ctx.JSON(http.StatusOK, ret)
}

func (s *apiServer) CreateEvent(ctx *gin.Context) {
	var req models.CreateEventRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	semesterId, err := uuid.Parse(req.SemesterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid semester ID specified in request"))
		return
	}

	event := models.Event{
		Name:             req.Name,
		Format:           req.Format,
		Notes:            req.Notes,
		SemesterID:       semesterId,
		StartDate:        req.StartDate,
		State:            models.EventStateStarted,
		StructureID:      req.StructureID,
		Rebuys:           0,
		PointsMultiplier: req.PointsMultiplier,
	}

	if err := s.store.Events().Create(&event); err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, event)
}

func (s *apiServer) GetEvent(ctx *gin.Context) {
	eventId, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil || eventId < 0 {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	event, err := s.store.Events().FindByID(int32(eventId))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (s *apiServer) UpdateEvent(ctx *gin.Context) {
	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil || eventID < 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	var req models.UpdateEventRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewEventService(s.store)
	event, err := svc.UpdateEvent(int32(eventID), &req)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (s *apiServer) UndoEndEvent(ctx *gin.Context) {
	eventId, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil || eventId < 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	svc := services.NewEventService(s.store)
	err = svc.UndoEndEvent(int32(eventId))
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusNoContent, "")
}

func (s *apiServer) EndEvent(ctx *gin.Context) {
	eventId, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil || eventId < 0 {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	svc := services.NewEventService(s.store)
	err = svc.EndEvent(int32(eventId))
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusNoContent, "")
}

func (s *apiServer) NewRebuy(ctx *gin.Context) {
	eventId, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil || eventId < 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
		return
	}

	svc := services.NewEventService(s.store)
	err = svc.NewRebuy(int32(eventId))
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusOK, "")
}
