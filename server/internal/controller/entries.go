package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type entriesController struct {
	db *gorm.DB
}

func NewEntriesController(db *gorm.DB) Controller {
	return &entriesController{db: db}
}

func (c *entriesController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters/:semesterId/events/:eventId/entries", middleware.UseAuthentication(c.db))
	group.POST("", middleware.UseAuthorization(c.db, "event.participant.create"), c.createEntry)
	group.GET("", middleware.UseAuthorization(c.db, "event.participant.list"), c.listEntries)
	group.POST(":entryId/sign-out", middleware.UseAuthorization(c.db, "event.participant.signout"), c.signOutEntry)
	group.POST(":entryId/sign-in", middleware.UseAuthorization(c.db, "event.participant.signin"), c.signInEntry)
	group.DELETE(":entryId", middleware.UseAuthorization(c.db, "event.participant.delete"), c.deleteEntry)
}

// createEntry handles the creation of a new participant entry for an event.
// It expects a CreateParticipantRequest in the request body and returns the created Participant.
//
// @Summary Create Entry
// @Description Create a new participant entry for an event
// @Tags Entries
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param entry body CreateParticipantRequest true "Entry data"
// @Success 201 {object} Participant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries [post]
func (c *entriesController) createEntry(ctx *gin.Context) {
	// Validate semester ID
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterId),
			),
		)
		return
	}

	// Validate event ID
	eventId := ctx.Param("eventId")
	if _, err := strconv.ParseInt(eventId, 10, 32); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", eventId),
			),
		)
		return
	}

	// Bind request body
	var req models.CreateParticipantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	// Create participant
	svc := services.NewParticipantsService(c.db)
	participant, err := svc.CreateParticipant(&req)
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

	ctx.JSON(http.StatusCreated, participant)
}

// listEntries handles the retrieval of all participant entries for a specific event.
// It returns a list of participants for the given event ID.
//
// @Summary List Entries
// @Description Retrieve a list of participant entries for a specific event
// @Tags Entries
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Success 200 {array} ListParticipantsResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries [get]
func (c *entriesController) listEntries(ctx *gin.Context) {
	// Validate semester ID
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterId),
			),
		)
		return
	}

	// Validate event ID
	eventId := ctx.Param("eventId")
	eventIdInt, err := strconv.ParseInt(eventId, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", eventId),
			),
		)
		return
	}

	// List participants
	svc := services.NewParticipantsService(c.db)
	participants, err := svc.ListParticipants(int32(eventIdInt))
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

	ctx.JSON(http.StatusOK, participants)
}

// signOutEntry handles signing out a participant from an event.
// It expects the entry ID (membership ID) in the URL path.
//
// @Summary Sign Out Entry
// @Description Sign out a participant from an event
// @Tags Entries
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param entryId path string true "Entry ID (Membership ID)"
// @Success 200 {object} Participant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries/{entryId}/sign-out [post]
func (c *entriesController) signOutEntry(ctx *gin.Context) {
	// Validate semester ID
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterId),
			),
		)
		return
	}

	// Validate event ID
	eventId := ctx.Param("eventId")
	eventIdInt, err := strconv.ParseInt(eventId, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", eventId),
			),
		)
		return
	}

	// Validate entry ID (membership ID)
	entryId := ctx.Param("entryId")
	membershipId, err := uuid.Parse(entryId)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Entry ID '%s' is not a valid UUID", entryId),
			),
		)
		return
	}

	// Update participant
	svc := services.NewParticipantsService(c.db)
	req := models.UpdateParticipantRequest{
		MembershipID: membershipId,
		EventID:      int32(eventIdInt),
		SignOut:      true,
		SignIn:       false,
	}

	participant, err := svc.UpdateParticipant(&req)
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

	ctx.JSON(http.StatusOK, participant)
}

// signInEntry handles signing in a participant to an event.
// It expects the entry ID (membership ID) in the URL path.
//
// @Summary Sign In Entry
// @Description Sign in a participant to an event
// @Tags Entries
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param entryId path string true "Entry ID (Membership ID)"
// @Success 200 {object} Participant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries/{entryId}/sign-in [post]
func (c *entriesController) signInEntry(ctx *gin.Context) {
	// Validate semester ID
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterId),
			),
		)
		return
	}

	// Validate event ID
	eventId := ctx.Param("eventId")
	eventIdInt, err := strconv.ParseInt(eventId, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", eventId),
			),
		)
		return
	}

	// Validate entry ID (membership ID)
	entryId := ctx.Param("entryId")
	membershipId, err := uuid.Parse(entryId)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Entry ID '%s' is not a valid UUID", entryId),
			),
		)
		return
	}

	// Update participant
	svc := services.NewParticipantsService(c.db)
	req := models.UpdateParticipantRequest{
		MembershipID: membershipId,
		EventID:      int32(eventIdInt),
		SignIn:       true,
		SignOut:      false,
	}

	participant, err := svc.UpdateParticipant(&req)
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

	ctx.JSON(http.StatusOK, participant)
}

// deleteEntry handles the deletion of a participant entry from an event.
// It expects the entry ID (membership ID) in the URL path.
//
// @Summary Delete Entry
// @Description Delete a participant entry from an event
// @Tags Entries
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param entryId path string true "Entry ID (Membership ID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries/{entryId} [delete]
func (c *entriesController) deleteEntry(ctx *gin.Context) {
	// Validate semester ID
	semesterId := ctx.Param("semesterId")
	if _, err := uuid.Parse(semesterId); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterId),
			),
		)
		return
	}

	// Validate event ID
	eventId := ctx.Param("eventId")
	eventIdInt, err := strconv.ParseInt(eventId, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Event ID '%s' is not a valid integer", eventId),
			),
		)
		return
	}

	// Validate entry ID (membership ID)
	entryId := ctx.Param("entryId")
	membershipId, err := uuid.Parse(entryId)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Entry ID '%s' is not a valid UUID", entryId),
			),
		)
		return
	}

	// Delete participant
	svc := services.NewParticipantsService(c.db)
	req := models.DeleteParticipantRequest{
		MembershipID: membershipId,
		EventID:      int32(eventIdInt),
	}

	err = svc.DeleteParticipant(&req)
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
