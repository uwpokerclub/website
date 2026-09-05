package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type eventClockController struct {
	store store.Store
}

// NewEventClockController creates a new instance of the event clock
// controller with the provided data store.
func NewEventClockController(st store.Store) Controller {
	return &eventClockController{store: st}
}

func (c *eventClockController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters/:semesterId/events/:eventId/clock", middleware.UseAuthentication(c.store))
	group.GET("", middleware.UseAuthorization("event.clock.get"), c.getClock)
	group.POST("pause", middleware.UseAuthorization("event.clock.control"), c.pauseClock)
	group.POST("resume", middleware.UseAuthorization("event.clock.control"), c.resumeClock)
	group.POST("adjust", middleware.UseAuthorization("event.clock.control"), c.adjustClock)
	group.POST("level", middleware.UseAuthorization("event.clock.control"), c.setClockLevel)
}

// findEvent validates the semester and event IDs from the URL path and
// returns the event, writing the appropriate error response and returning ok
// = false if validation or lookup fails.
func (c *eventClockController) findEvent(ctx *gin.Context) (event models.Event, ok bool) {
	semesterID, err := uuid.Parse(ctx.Param("semesterId"))
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", ctx.Param("semesterId")),
			),
		)
		return models.Event{}, false
	}

	eventID, err := strconv.ParseInt(ctx.Param("eventId"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", ctx.Param("eventId")),
			),
		)
		return models.Event{}, false
	}

	event, err = c.store.Events().FindBySemesterAndID(semesterID, int32(eventID))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(
				http.StatusNotFound,
				apierrors.NotFound(
					fmt.Sprintf("Event '%d' not found for semester '%s'", eventID, semesterID),
				),
			)
			return models.Event{}, false
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return models.Event{}, false
	}

	return event, true
}

// eventEndedMessage is shared by requireNotEnded's fast-path rejection and
// handleClockError's authoritative one, so both routes to the same 409 read
// identically.
const eventEndedMessage = "This event has ended. Its clock cannot be controlled."

// requireNotEnded aborts the request with 409 if the event has ended. This is
// a fast path only, checked before opening a transaction; it is not the
// authoritative check. Only control actions call this; reading the clock of
// an ended event is allowed. The service re-checks state itself inside its
// own transaction (surfaced as services.ErrEventEnded via handleClockError),
// since this read can go stale between here and that transaction if the
// event ends concurrently.
func (c *eventClockController) requireNotEnded(ctx *gin.Context, event models.Event) bool {
	if event.State == models.EventStateEnded {
		ctx.AbortWithStatusJSON(http.StatusConflict, apierrors.Conflict(eventEndedMessage))
		return false
	}
	return true
}

// withActiveEvent resolves the event from the URL path and rejects it with
// 409 if it has ended, returning ok = false in either case after writing the
// appropriate response. It is the only path a control action has to an
// event, so a future control action cannot copy-paste its way past the
// ended-event check the way it could if each handler re-inlined
// findEvent+requireNotEnded itself.
func (c *eventClockController) withActiveEvent(ctx *gin.Context) (event models.Event, ok bool) {
	event, ok = c.findEvent(ctx)
	if !ok {
		return models.Event{}, false
	}
	if !c.requireNotEnded(ctx, event) {
		return models.Event{}, false
	}
	return event, true
}

// respondClock writes a derived clock as the standard ClockState response.
func (c *eventClockController) respondClock(ctx *gin.Context, derived models.DerivedClock) {
	ctx.JSON(http.StatusOK, models.NewClockState(derived, time.Now().UTC()))
}

// handleClockError maps a services.EventClockService sentinel error to the
// appropriate HTTP response.
func (c *eventClockController) handleClockError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrEmptyStructure):
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			apierrors.NotFound("This event's structure has no blind levels; it has no clock."),
		)
	case errors.Is(err, services.ErrInvalidLevel):
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
	case errors.Is(err, services.ErrEventEnded):
		ctx.AbortWithStatusJSON(http.StatusConflict, apierrors.Conflict(eventEndedMessage))
	default:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
	}
}

// getClock retrieves the current, fully derived state of an event's
// tournament clock, lazily creating it on first read.
//
// @Summary Get Event Clock
// @Description Retrieve the current, fully derived state of an event's tournament clock
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} ClockState
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/clock [get]
func (c *eventClockController) getClock(ctx *gin.Context) {
	event, ok := c.findEvent(ctx)
	if !ok {
		return
	}

	svc := services.NewEventClockService(c.store)
	derived, err := svc.GetClock(event.ID)
	if err != nil {
		c.handleClockError(ctx, err)
		return
	}

	c.respondClock(ctx, derived)
}

// pauseClock freezes the clock at its current, rolled-forward remaining
// time. Pausing an already-paused clock is a no-op.
//
// @Summary Pause Event Clock
// @Description Freeze an event's tournament clock at its current remaining time
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} ClockState
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/clock/pause [post]
func (c *eventClockController) pauseClock(ctx *gin.Context) {
	event, ok := c.withActiveEvent(ctx)
	if !ok {
		return
	}

	svc := services.NewEventClockService(c.store)
	derived, err := svc.Pause(event.ID)
	if err != nil {
		c.handleClockError(ctx, err)
		return
	}

	c.respondClock(ctx, derived)
}

// resumeClock unfreezes the clock, restoring exactly the remaining time it
// had when paused. Resuming an already-running clock is a no-op.
//
// @Summary Resume Event Clock
// @Description Unfreeze an event's tournament clock, restoring its remaining time
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} ClockState
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/clock/resume [post]
func (c *eventClockController) resumeClock(ctx *gin.Context) {
	event, ok := c.withActiveEvent(ctx)
	if !ok {
		return
	}

	svc := services.NewEventClockService(c.store)
	derived, err := svc.Resume(event.ID)
	if err != nil {
		c.handleClockError(ctx, err)
		return
	}

	c.respondClock(ctx, derived)
}

// adjustClock shifts the current level's deadline by deltaSeconds, which may
// be negative and roll into the next level.
//
// @Summary Adjust Event Clock
// @Description Shift the current level's deadline by a number of seconds, which may be negative
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param body body AdjustClockRequest true "Adjustment"
// @Success 200 {object} ClockState
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/clock/adjust [post]
func (c *eventClockController) adjustClock(ctx *gin.Context) {
	event, ok := c.withActiveEvent(ctx)
	if !ok {
		return
	}

	var req models.AdjustClockRequest
	if !BindJSON(ctx, &req) {
		return
	}

	svc := services.NewEventClockService(c.store)
	derived, err := svc.Adjust(event.ID, req.DeltaSeconds)
	if err != nil {
		c.handleClockError(ctx, err)
		return
	}

	c.respondClock(ctx, derived)
}

// setClockLevel jumps the clock to an absolute level index with a full,
// fresh duration for that level.
//
// @Summary Set Event Clock Level
// @Description Jump the clock to an absolute level index with a full, fresh duration
// @Tags Events
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param body body SetClockLevelRequest true "Level index"
// @Success 200 {object} ClockState
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/clock/level [post]
func (c *eventClockController) setClockLevel(ctx *gin.Context) {
	event, ok := c.withActiveEvent(ctx)
	if !ok {
		return
	}

	var req models.SetClockLevelRequest
	if !BindJSON(ctx, &req) {
		return
	}

	svc := services.NewEventClockService(c.store)
	derived, err := svc.SetLevel(event.ID, *req.Index)
	if err != nil {
		c.handleClockError(ctx, err)
		return
	}

	c.respondClock(ctx, derived)
}
