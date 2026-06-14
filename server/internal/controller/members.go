package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/store"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type membersController struct {
	store store.Store
	db    *gorm.DB
}

// NewMembersController creates a new instance of membersController
func NewMembersController(db *gorm.DB, st store.Store) Controller {
	return &membersController{
		store: st,
		db:    db,
	}
}

func (c *membersController) LoadRoutes(router *gin.RouterGroup) {
	members := router.Group("members", middleware.UseAuthentication(c.db))
	members.POST("", middleware.UseAuthorization("user.create"), c.createMember)
	members.GET("", middleware.UseAuthorization("user.list"), c.listMembers)
	members.GET("/:id", middleware.UseAuthorization("user.get"), c.getMember)
	members.PATCH("/:id", middleware.UseAuthorization("user.edit"), c.updateMember)
	members.DELETE("/:id", middleware.UseAuthorization("user.delete"), c.deleteMember)
}

func validateMemberID(ctx *gin.Context) (uint64, error) {
	idParam := ctx.Param("id")
	idInt, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, apierrors.InvalidRequest(
			fmt.Sprintf("Member ID '%s' is not a valid integer", idParam),
		)
	}

	if idInt <= 0 {
		return 0, apierrors.InvalidRequest("Member ID must be a positive integer")
	}

	return uint64(idInt), nil
}

// createMember handles the creation of a new Member
//
// @Summary Create a new Member
// @Description Create a new Member with the provided details
// @Tags Members
// @Accept json
// @Produce json
// @Param member body CreateMemberRequest true "Member details"
// @Success 201 {object} Member
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /members [post]
func (c *membersController) createMember(ctx *gin.Context) {
	var req models.CreateUserRequest
	if !BindJSON(ctx, &req) {
		return
	}

	member := models.User{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Faculty:   req.Faculty,
		QuestID:   req.QuestID,
	}

	if err := c.store.Members().Create(&member); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusCreated, member)
}

// parseListMembersQueryParams parses query parameters for listing members
func (c *membersController) parseListMembersQueryParams(ctx *gin.Context) *models.ListUsersFilter {
	filter := &models.ListUsersFilter{}

	if idStr, exists := ctx.GetQuery("id"); exists {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			filter.ID = &id
		}
	}

	if email, exists := ctx.GetQuery("email"); exists {
		filter.Email = &email
	}

	if name, exists := ctx.GetQuery("name"); exists {
		filter.Name = &name
	}

	if faculty, exists := ctx.GetQuery("faculty"); exists {
		filter.Faculty = &faculty
	}

	return filter
}

// listMembers handles listing Members with optional filters
//
// @Summary List Members
// @Description Retrieve a list of Members with optional filters
// @Tags Members
// @Accept json
// @Produce json
// @Param id query int false "Filter by Member ID"
// @Param email query string false "Filter by Member Email"
// @Param name query string false "Filter by Member Name"
// @Param faculty query string false "Filter by Member Faculty"
// @Success 200 {array} Member
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /members [get]
func (c *membersController) listMembers(ctx *gin.Context) {
	filter := c.parseListMembersQueryParams(ctx)

	pagination, err := models.ParsePagination(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	members, total, err := c.store.Members().List(filter, &pagination)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, models.ListResponse[*models.User]{
		Data:  members,
		Total: total,
	})
}

// getMember handles retrieving a Member by ID
//
// @Summary Get Member by ID
// @Description Retrieve a Member by their ID
// @Tags Members
// @Accept json
// @Produce json
// @Param id path int true "Member ID"
// @Success 200 {object} Member
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /members/{id} [get]
func (c *membersController) getMember(ctx *gin.Context) {
	memberID, err := validateMemberID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	member, err := c.store.Members().FindByID(memberID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(
				http.StatusNotFound,
				apierrors.NotFound(fmt.Sprintf("Member with ID %d not found", memberID)),
			)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, member)
}

// updateMember handles updating a Member by ID
//
// @Summary Update Member by ID
// @Description Update a Member's details by their ID
// @Tags Members
// @Accept json
// @Produce json
// @Param id path int true "Member ID"
// @Param member body UpdateMemberRequest true "Updated Member details"
// @Success 200 {object} Member
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /members/{id} [patch]
func (c *membersController) updateMember(ctx *gin.Context) {
	memberID, err := validateMemberID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	var req models.UpdateUserRequest
	if !BindJSON(ctx, &req) {
		return
	}

	tx, err := c.store.BeginTx()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}
	defer tx.Rollback()

	member, err := tx.Members().FindByID(memberID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(
				http.StatusNotFound,
				apierrors.NotFound(fmt.Sprintf("Member with ID %d not found", memberID)),
			)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	if req.FirstName != "" {
		member.FirstName = req.FirstName
	}
	if req.LastName != "" {
		member.LastName = req.LastName
	}
	if req.Email != "" {
		member.Email = req.Email
	}
	if req.Faculty != "" {
		member.Faculty = req.Faculty
	}
	if req.QuestID != "" {
		member.QuestID = req.QuestID
	}

	if err := tx.Members().Update(member); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	if err := tx.Commit(); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, member)
}

// deleteMember handles deleting a Member by ID
//
// @Summary Delete Member by ID
// @Description Delete a Member by their ID
// @Tags Members
// @Accept json
// @Produce json
// @Param id path int true "Member ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /members/{id} [delete]
func (c *membersController) deleteMember(ctx *gin.Context) {
	memberID, err := validateMemberID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	err = c.store.Members().Delete(memberID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}
