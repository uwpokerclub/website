package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *apiServer) ListMemberships(ctx *gin.Context) {
	semesterId, err := uuid.Parse(ctx.Query("semesterId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid semester UUID specified in query."))
	}

	svc := services.NewMembershipService(s.db)
	memberships, err := svc.ListMemberships(semesterId)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

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
		ctx.JSON(err.(e.APIErrorResponse).Code, e.InvalidRequest(err.Error()))
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
