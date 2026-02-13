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

// NewEntriesController creates a new instance of the entries controller
// with the provided database connection.
func NewEntriesController(db *gorm.DB) Controller {
	return &entriesController{db: db}
}

func (c *entriesController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters/:semesterId/events/:eventId/entries", middleware.UseAuthentication(c.db))
	group.POST("", middleware.UseAuthorization("event.participant.create"), c.createEntry)
	group.GET("", middleware.UseAuthorization("event.participant.list"), c.listEntries)
	group.POST(":entryId/sign-out", middleware.UseAuthorization("event.participant.signout"), c.signOutEntry)
	group.POST(":entryId/sign-in", middleware.UseAuthorization("event.participant.signin"), c.signInEntry)
	group.DELETE(":entryId", middleware.UseAuthorization("event.participant.delete"), c.deleteEntry)
}

// validateSemesterID validates and returns the semester UUID from the path parameter.
// It returns uuid.Nil and an API error if validation fails.
func (c *entriesController) validateSemesterID(ctx *gin.Context) (uuid.UUID, error) {
	semesterID := ctx.Param("semesterId")
	semesterUUID, err := uuid.Parse(semesterID)
	if err != nil {
		return uuid.Nil, apierrors.InvalidRequest(
			fmt.Sprintf("Semester ID '%s' is not a valid UUID", semesterID),
		)
	}
	return semesterUUID, nil
}

// validateEventID validates and returns the event ID as int32 from the path parameter.
// It returns 0 and an API error if validation fails.
func (c *entriesController) validateEventID(ctx *gin.Context) (int32, error) {
	eventIDStr := ctx.Param("eventId")
	eventIDInt, err := strconv.ParseInt(eventIDStr, 10, 32)
	if err != nil {
		return 0, apierrors.InvalidRequest(
			fmt.Sprintf("Event ID '%s' is not a valid integer", eventIDStr),
		)
	}
	return int32(eventIDInt), nil
}

// createEntry handles the creation of new participant entries for an event.
// It expects an array of membership UUIDs in the request body and returns the results.
//
// @Summary Create Entries
// @Description Create new participant entries for an event
// @Tags Entries
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param eventId path string true "Event ID"
// @Param membershipIds body []string true "Array of membership UUIDs"
// @Success 207 {array} CreateEntryResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries [post]
func (c *entriesController) createEntry(ctx *gin.Context) {
	// Validate semester ID
	if _, err := c.validateSemesterID(ctx); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate event ID
	eventID, err := c.validateEventID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Bind request body - array of UUID strings
	var membershipIdStrs []string
	if err := ctx.ShouldBindJSON(&membershipIdStrs); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	// Validate array is not empty
	if len(membershipIdStrs) == 0 {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest("Array of membership IDs cannot be empty"),
		)
		return
	}

	// Parse membership UUIDs
	membershipIds := make([]uuid.UUID, 0, len(membershipIdStrs))
	for _, idStr := range membershipIdStrs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			ctx.AbortWithStatusJSON(
				http.StatusBadRequest,
				apierrors.InvalidRequest(
					fmt.Sprintf("Invalid UUID format: '%s'", idStr),
				),
			)
			return
		}
		membershipIds = append(membershipIds, id)
	}

	// Create participants and collect results
	svc := services.NewParticipantsService(c.db)
	results := make([]models.CreateEntryResult, 0, len(membershipIds))

	for _, membershipId := range membershipIds {
		req := models.CreateParticipantRequest{
			MembershipID: membershipId,
			EventID:      eventID,
		}

		participant, err := svc.CreateParticipant(&req)
		if err != nil {
			// Collect error but continue processing
			errMsg := err.Error()
			if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
				errMsg = apiErr.Message
			}
			results = append(results, models.CreateEntryResult{
				MembershipID: membershipId,
				Status:       "error",
				Error:        errMsg,
			})
		} else {
			// Success
			results = append(results, models.CreateEntryResult{
				MembershipID: membershipId,
				Status:       "created",
				Participant:  participant,
			})
		}
	}

	ctx.JSON(http.StatusMultiStatus, results)
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
// @Success 200 {array} Participant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries [get]
func (c *entriesController) listEntries(ctx *gin.Context) {
	// Validate semester ID (validates URL structure but not used by service layer)
	if _, err := c.validateSemesterID(ctx); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate event ID
	eventID, err := c.validateEventID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	pagination, err := models.ParsePagination(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	// List participants
	svc := services.NewParticipantsService(c.db)
	participants, total, err := svc.ListParticipantsV2(eventID, &pagination)
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

	ctx.JSON(http.StatusOK, models.ListResponse[models.Participant]{
		Data:  participants,
		Total: total,
	})
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
// @Param entryId path string true "Membership ID (UUID format)"
// @Success 200 {object} Participant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries/{entryId}/sign-out [post]
func (c *entriesController) signOutEntry(ctx *gin.Context) {
	// Validate semester ID
	if _, err := c.validateSemesterID(ctx); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate event ID
	eventID, err := c.validateEventID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate entry ID (membership ID)
	entryID := ctx.Param("entryId")
	membershipID, err := uuid.Parse(entryID)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Entry ID '%s' is not a valid UUID", entryID),
			),
		)
		return
	}

	// Update participant
	svc := services.NewParticipantsService(c.db)
	req := models.UpdateParticipantRequest{
		MembershipID: membershipID,
		EventID:      eventID,
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
// @Param entryId path string true "Membership ID (UUID format)"
// @Success 200 {object} Participant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries/{entryId}/sign-in [post]
func (c *entriesController) signInEntry(ctx *gin.Context) {
	// Validate semester ID
	if _, err := c.validateSemesterID(ctx); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate event ID
	eventID, err := c.validateEventID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate entry ID (membership ID)
	entryID := ctx.Param("entryId")
	membershipID, err := uuid.Parse(entryID)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Entry ID '%s' is not a valid UUID", entryID),
			),
		)
		return
	}

	// Update participant
	svc := services.NewParticipantsService(c.db)
	req := models.UpdateParticipantRequest{
		MembershipID: membershipID,
		EventID:      eventID,
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
// @Param entryId path string true "Membership ID (UUID format)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/events/{eventId}/entries/{entryId} [delete]
func (c *entriesController) deleteEntry(ctx *gin.Context) {
	// Validate semester ID (validates URL structure but not used by service layer)
	if _, err := c.validateSemesterID(ctx); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate event ID
	eventID, err := c.validateEventID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Validate entry ID (membership ID)
	entryID := ctx.Param("entryId")
	membershipID, err := uuid.Parse(entryID)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Entry ID '%s' is not a valid UUID", entryID),
			),
		)
		return
	}

	// Delete participant
	svc := services.NewParticipantsService(c.db)
	req := models.DeleteParticipantRequest{
		MembershipID: membershipID,
		EventID:      eventID,
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
