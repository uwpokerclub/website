package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *apiServer) ListParticipants(ctx *gin.Context) {
	eventId, err := strconv.ParseInt(ctx.Query("eventId"), 10, 32)
	if err != nil || eventId < 0 {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID in query"))
		return
	}

	participants, _, err := s.store.Entries().List(&models.ListParticipantsFilter{EventID: int32(eventId)})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	// Preserves the field mapping of the original raw-SQL query this replaces: "id" is the
	// linked user's numeric ID (not the participant/entry ID), "membershipId" is the
	// membership's UUID.
	ret := make([]models.ListParticipantsResult, 0, len(participants))
	for _, p := range participants {
		if p.Membership == nil || p.Membership.User == nil {
			continue
		}
		ret = append(ret, models.ListParticipantsResult{
			ID:           int32(p.Membership.User.ID),
			MembershipId: p.Membership.ID,
			FirstName:    p.Membership.User.FirstName,
			LastName:     p.Membership.User.LastName,
			SignedOutAt:  p.SignedOutAt,
			Placement:    p.Placement,
		})
	}

	ctx.JSON(http.StatusOK, ret)
}

func (s *apiServer) CreateParticipant(ctx *gin.Context) {
	var req models.CreateParticipantRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewParticipantsService(s.store)
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

	svc := services.NewParticipantsService(s.store)
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

	svc := services.NewParticipantsService(s.store)
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

	err = s.store.Entries().Delete(req.MembershipID, req.EventID)
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(http.StatusNotFound, e.NotFound("Entry not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.String(http.StatusNoContent, "")
}
