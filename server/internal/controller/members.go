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
	"gorm.io/gorm"
)

type membersController struct {
	db *gorm.DB
}

// NewMembersController creates a new instance of membersController
func NewMembersController(db *gorm.DB) *membersController {
	return &membersController{db: db}
}

func (c *membersController) LoadRoutes(router *gin.RouterGroup) {
	members := router.Group("members", middleware.UseAuthentication(c.db))
	members.POST("", middleware.UseAuthorization(c.db, "user.create"), c.createMember)
	members.GET("", middleware.UseAuthorization(c.db, "user.list"), c.listMembers)
	members.GET("/:id", middleware.UseAuthorization(c.db, "user.get"), c.getMember)
	members.PATCH("/:id", middleware.UseAuthorization(c.db, "user.edit"), c.updateMember)
	members.DELETE("/:id", middleware.UseAuthorization(c.db, "user.delete"), c.deleteMember)
}

func validateMemberID(ctx *gin.Context) (uint64, error) {
	idParam := ctx.Param("id")
	idInt, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, apierrors.InvalidRequest(
			fmt.Sprintf("Member ID '%s' is not a valid positive integer", idParam),
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewUserService(c.db)
	member, err := svc.CreateUser(&req)
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

	ctx.JSON(http.StatusCreated, member)
}

// parseListMembersQueryParams parses query parameters for listing members
func (c *membersController) parseListMembersQueryParams(ctx *gin.Context) *models.ListUsersFilter {
	filter := &models.ListUsersFilter{}

	// Parse ID filter
	if idStr, exists := ctx.GetQuery("id"); exists {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			filter.ID = &id
		}
	}

	// Parse email filter
	if email, exists := ctx.GetQuery("email"); exists {
		filter.Email = &email
	}

	// Parse name filter
	if name, exists := ctx.GetQuery("name"); exists {
		filter.Name = &name
	}

	// Parse faculty filter
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

	svc := services.NewUserService(c.db)
	members, err := svc.ListUsers(filter)
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

	ctx.JSON(http.StatusOK, members)
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

	svc := services.NewUserService(c.db)
	member, err := svc.GetUser(memberID)
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewUserService(c.db)
	member, err := svc.UpdateUser(memberID, &req)
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

	svc := services.NewUserService(c.db)
	err = svc.DeleteUser(memberID)
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
