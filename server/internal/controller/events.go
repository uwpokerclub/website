package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type eventsController struct {
	db *gorm.DB
}

func NewEventsController(db *gorm.DB) Controller {
	return &eventsController{db: db}
}

func (s *eventsController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters/:semesterId/events", middleware.UseAuthentication(s.db))
	group.POST("", middleware.UseAuthorization(s.db, "event.create"), s.createEvent)
	group.GET("", middleware.UseAuthorization(s.db, "event.list"), s.listEvents)
	group.GET(":eventId", middleware.UseAuthorization(s.db, "event.get"), s.getEvent)
	group.PATCH(":eventId", middleware.UseAuthorization(s.db, "event.edit"), s.updateEvent)
	group.POST(":eventId/end", middleware.UseAuthorization(s.db, "event.end"), s.endEvent)
	group.POST(
		":eventId/restart",
		middleware.UseAuthorization(s.db, "event.restart"),
		s.restartEvent,
	)
	group.POST(":eventId/rebuy", middleware.UseAuthorization(s.db, "event.rebuy"), s.rebuyEvent)
}

// createEvent handles the creation of a new event.
// It expects a CreateEventRequest in the request body and returns the created Event.
//
// @Summary Create Event
// @Description Create a new event
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param event body CreateEventRequest true "Event data"
// @Success 201 {object} Event
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events [post]
func (s *eventsController) createEvent(ctx *gin.Context) {
	// Retrieve the semester ID from the URL path and parse it as a UUID
	semesterId := ctx.Param("semesterId")
	semesterUUID, err := uuid.Parse(semesterId)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterId),
			),
		)
		return
	}

	// Retrieve the request body and bind it to CreateEventRequest
	var req models.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	// Initialize the event service and create the event
	svc := services.NewEventService(s.db)
	event, err := svc.CreateEventV2(semesterUUID, &req)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusCreated, event)
}

// listEvents handles the retrieval of events for a specific semester.
// It returns a list of events for the given semester ID.
// It expects the semester ID in the URL path.
//
// @Summary List Events
// @Description Retrieve a list of events for a specific semester
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {array} Event
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events [get]
func (s *eventsController) listEvents(ctx *gin.Context) {
	// Retrieve the semester ID from the URL path and parse it as a UUID
	semesterID, err := uuid.Parse(ctx.Param("semesterId"))
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return
	}

	// Initialize the event service and list events for the semester
	svc := services.NewEventService(s.db)
	events, err := svc.ListEventsV2(semesterID)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, events)
}

// getEvent handles the retrieval of a specific event by its ID.
// It expects the semester ID in the URL path and the event ID as a URL parameter.
//
// @Summary Get Event
// @Description Retrieve a specific event by its ID
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} Event
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId} [get]
func (s *eventsController) getEvent(ctx *gin.Context) {
	// Retrieve the semester ID from the URL path and parse it as a UUID
	semesterID, err := uuid.Parse(ctx.Param("semesterId"))
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return
	}

	// Retrieve the event ID from the URL parameter and parse it as a integer
	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", ctx.Param("eventId")),
			),
		)
		return
	}

	// Initialize the event service and get the event by ID
	svc := services.NewEventService(s.db)
	event, err := svc.GetEventByID(semesterID, int32(eventID))
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	if event == nil {
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			apierrors.NotFound(
				fmt.Sprintf("Event '%d' not found for semester '%s'", eventID, semesterID),
			),
		)
		return
	}

	ctx.JSON(http.StatusOK, event)
}

// updateEvent handles the editing of an existing event.
// It expects the semester ID in the URL path and the event ID as a URL parameter.
//
// @Summary Update Event
// @Description Update an existing event
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param event body UpdateEventRequestV2 true "Event data"
// @Success 200 {object} Event
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId} [patch]
func (s *eventsController) updateEvent(ctx *gin.Context) {
	semesterID, err := uuid.Parse(ctx.Param("semesterId"))
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return
	}

	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", ctx.Param("eventId")),
			),
		)
		return
	}

	// Parse request body once into a map for partial update handling
	requestValues := make(map[string]any)
	if err := ctx.ShouldBindJSON(&requestValues); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(fmt.Sprintf("Error parsing request body: %s", err.Error())),
		)
		return
	}

	updateMap, err := s.validateAndCreateEventUpdateMap(requestValues)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Error converting request to update map: %s", err.Error()),
			),
		)
		return
	}

	svc := services.NewEventService(s.db)

	event, err := svc.GetEventByID(semesterID, int32(eventID))
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	if event == nil {
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			apierrors.NotFound(
				fmt.Sprintf("Event '%d' not found for semester '%s'", eventID, semesterID),
			),
		)
		return
	}

	err = svc.UpdateEventV2(event, updateMap)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (controller *eventsController) validateAndCreateEventUpdateMap(
	requestValues map[string]any,
) (map[string]any, error) {
	updateMap := make(map[string]any)

	for key, value := range requestValues {
		switch key {
		case "name":
			if value == nil {
				return nil, errors.New("name cannot be null")
			}
			strValue, ok := value.(string)
			if !ok {
				return nil, errors.New("name must be a string")
			}
			if strValue == "" {
				return nil, errors.New("name must be a non-empty string")
			}
			updateMap["name"] = strValue
		case "format":
			if value == nil {
				return nil, errors.New("format cannot be null")
			}
			strValue, ok := value.(string)
			if !ok {
				return nil, errors.New("format must be a string")
			}
			if strValue == "" {
				return nil, errors.New("format must be a non-empty string")
			}
			updateMap["format"] = strValue
		case "notes":
			if value != nil {
				strValue, ok := value.(string)
				if !ok {
					return nil, errors.New("notes must be a string")
				}
				updateMap["notes"] = strValue
			} else {
				updateMap["notes"] = nil
			}
		case "startDate":
			if value == nil {
				return nil, errors.New("startDate cannot be null")
			}
			strValue, ok := value.(string)
			if !ok {
				return nil, errors.New("startDate must be a string")
			}
			tmpTime, err := time.Parse(time.RFC3339, strValue)
			if err != nil {
				return nil, errors.New("startDate must be in RFC3339 format")
			}
			updateMap["start_date"] = tmpTime
		case "pointsMultiplier":
			if value == nil {
				return nil, errors.New("pointsMultiplier cannot be null")
			}
			floatValue, ok := value.(float64)
			if !ok {
				return nil, errors.New("pointsMultiplier must be a number")
			}
			if floatValue < 0 {
				return nil, errors.New("pointsMultiplier must be a non-negative number")
			}
			updateMap["points_multiplier"] = float32(floatValue)
		default:
			return nil, fmt.Errorf(
				"failed to validate event update request: unknown field: %s",
				key,
			)
		}
	}

	return updateMap, nil
}

// endEvent handles the ending of an event.
// It expects the semester ID in the URL path and the event ID as a URL parameter.
//
// @Summary End Event
// @Description End an existing event
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/end [post]
func (s *eventsController) endEvent(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return
	}

	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", ctx.Param("eventId")),
			),
		)
		return
	}

	svc := services.NewEventService(s.db)
	if err := svc.EndEvent(int32(eventID)); err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// restartEvent handles the restarting of an event.
// It expects the semester ID in the URL path and the event ID as a URL parameter.
//
// @Summary Restart Event
// @Description Restart an existing Event
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/restart [post]
func (s *eventsController) restartEvent(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return
	}

	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", ctx.Param("eventId")),
			),
		)
		return
	}

	svc := services.NewEventService(s.db)
	if err := svc.UndoEndEvent(int32(eventID)); err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// rebuyEvent handles adding a new rebuy entry to an event.
// It expects the semester ID in the URL path and the event ID as a URL parameter.
//
// @Summary Rebuy Event
// @Description Add a new rebuy entry to an event
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/rebuy [post]
func (s *eventsController) rebuyEvent(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return
	}
	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", ctx.Param("eventId")),
			),
		)
		return
	}

	svc := services.NewEventService(s.db)
	err = svc.NewRebuy(int32(eventID))
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}
