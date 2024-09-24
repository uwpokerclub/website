package server

import (
	"strconv"

	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListMemberships returns an array of all members.
//
// This handler supports the following pagination query parameters:
//
//	offset: number - number of members to offset the list by
//	limit: number - the amount of members tor return in the response
//
// This handler supports the following filters
//
//	semesterId: uuid - the ID of the semester to get members for
func (s *apiServer) ListMemberships(ctx *gin.Context) {
	// Retrieve the semester ID from the query parameters
	semesterId, err := uuid.Parse(ctx.Query("semesterId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid semester UUID specified in query."))
		return
	}

	// Retrieve the limit pagination parameter from the query string
	var limit *int
	limitVal, err := strconv.Atoi(ctx.Query("limit"))
	if err == nil {
		// Ensure limit is within the valid values, if not default it to 100
		if limitVal < 0 || limitVal > 100 {
			limitVal = 100
		}
		limit = &limitVal
	}

	// Retrieve the offset pagination parameter from the query string
	var offset *int
	offsetVal, err := strconv.Atoi(ctx.Query("offset"))
	if err == nil {
		// Ensure offset is greater than 0, otherwise default to 0
		if offsetVal < 0 {
			offsetVal = 0
		}
		offset = &offsetVal
	}

	// Set up the filter
	filter := models.ListMembershipsFilter{
		Limit:      limit,
		Offset:     offset,
		SemesterID: &semesterId,
	}

	// Get the list fo all the members
	svc := services.NewMembershipService(s.db)
	memberships, err := svc.ListMemberships(&filter)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	// Return as JSON with status 200
	ctx.JSON(http.StatusOK, memberships)
}

func (s *apiServer) CreateMembership(ctx *gin.Context) {
	var req models.CreateMembershipRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewMembershipService(s.db)
	membership, err := svc.CreateMembership(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, membership)
}

func (s *apiServer) GetMembership(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid membership UUID specified in request."))
		return
	}

	svc := services.NewMembershipService(s.db)
	membership, err := svc.GetMembership(id)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, membership)
}

func (s *apiServer) UpdateMembership(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid membership UUID specified in request."))
		return
	}

	var req models.UpdateMembershipRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}
	req.ID = id

	svc := services.NewMembershipService(s.db)
	membership, err := svc.UpdateMembership(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, membership)
}
