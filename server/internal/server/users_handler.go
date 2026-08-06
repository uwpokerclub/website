package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"errors"
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

	user := models.User{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Faculty:   req.Faculty,
		QuestID:   req.QuestID,
	}

	if err := s.store.Members().Create(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
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

	users, _, err := s.store.Members().List(filter, &models.Pagination{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
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

	user, err := s.store.Members().FindByID(uint64(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
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

	tx, err := s.store.BeginTx()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}
	defer tx.Rollback()

	user, err := tx.Members().FindByID(uint64(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Faculty != "" {
		user.Faculty = req.Faculty
	}
	if req.QuestID != "" {
		user.QuestID = req.QuestID
	}

	if err := tx.Members().Update(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
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

	err = s.store.Members().Delete(uint64(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.String(http.StatusNoContent, "")
}
