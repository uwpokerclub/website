package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *apiServer) CreateUser(ctx *gin.Context) {
	var req models.CreateUserRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewUserService(s.db)
	user, err := svc.CreateUser(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func parseListUsersQueryParams(ctx *gin.Context) *models.ListUsersFilter {
	// Get filter parameters from query params
	filter := models.ListUsersFilter{}

	// Add user's ID to filter if it is present
	if stringID, exists := ctx.GetQuery("id"); exists {
		ID, err := strconv.ParseUint(stringID, 10, 64)
		if err == nil {
			filter.ID = &ID
		}
	}

	// Add user's email to filter if it is present
	if email, exists := ctx.GetQuery("email"); exists {
		filter.Email = &email
	}

	// Add user's name to filter if it is present
	if name, exists := ctx.GetQuery("name"); exists {
		filter.Name = &name
	}

	// Add user's faculty to filter if it is present {
	if faculty, exists := ctx.GetQuery("faculty"); exists {
		filter.Faculty = &faculty
	}

	return &filter
}

func (s *apiServer) ListUsers(ctx *gin.Context) {
	filter := parseListUsersQueryParams(ctx)

	svc := services.NewUserService(s.db)
	users, err := svc.ListUsers(filter)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (s *apiServer) GetUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
		return
	}

	svc := services.NewUserService(s.db)
	user, err := svc.GetUser(uint64(id))
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *apiServer) UpdateUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
		return
	}

	var req models.UpdateUserRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewUserService(s.db)
	user, err := svc.UpdateUser(uint64(id), &req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *apiServer) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
		return
	}

	svc := services.NewUserService(s.db)
	err = svc.DeleteUser(uint64(id))
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusNoContent, "")
}
