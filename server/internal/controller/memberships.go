package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type membershipsController struct {
	db *gorm.DB
}

// NewMembershipsController creates a new instance of membershipsController
func NewMembershipsController(db *gorm.DB) Controller {
	return &membershipsController{db: db}
}

func (c *membershipsController) LoadRoutes(router *gin.RouterGroup) {
	memberships := router.Group("semesters/:semesterId/memberships", middleware.UseAuthentication(c.db))
	memberships.POST("", middleware.UseAuthorization("membership.create"), c.createMembership)
	memberships.GET("", middleware.UseAuthorization("membership.list"), c.listMemberships)
	memberships.GET("/:id", middleware.UseAuthorization("membership.get"), c.getMembership)
	memberships.PATCH(
		"/:id",
		middleware.UseAuthorization("membership.edit"),
		c.updateMembership,
	)
}

func validateSemesterID(ctx *gin.Context) (uuid.UUID, error) {
	idParam := ctx.Param("semesterId")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return uuid.Nil, apierrors.InvalidRequest(
			"Semester ID '" + idParam + "' is not a valid UUID",
		)
	}

	return id, nil
}

func validateMembershipID(ctx *gin.Context) (uuid.UUID, error) {
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return uuid.Nil, apierrors.InvalidRequest(
			"Membership ID '" + idParam + "' is not a valid UUID",
		)
	}

	return id, nil
}

// createMembership handles the creation of a new Membership
//
// @Summary Create a new Membership
// @Description Create a new Membership with the provided details
// @Tags Memberships
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param membership body CreateMembershipRequest true "Membership details"
// @Success 201 {object} Membership
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/memberships [post]
func (c *membershipsController) createMembership(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	var req models.CreateMembershipRequestV2
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewMembershipService(c.db)
	membership, err := svc.CreateMembershipV2(semesterID, &req)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		// Check for validation errors
		errMsg := err.Error()
		if errMsg == "cannot create membership that is not paid and discounted" {
			ctx.AbortWithStatusJSON(
				http.StatusBadRequest,
				apierrors.InvalidRequest(errMsg),
			)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusCreated, membership)
}

func (c *membershipsController) parseListMembershipsQueryParams(
	ctx *gin.Context,
) *models.ListMembershipsFilter {
	// Implementation for parsing query parameters
	filter := &models.ListMembershipsFilter{}

	// Parse limit and offset filters
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = &limit
		}
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = &offset
		}
	}

	return filter
}

// listMemberships handles listing all memberships
//
// @Summary List all Memberships
// @Description Retrieve a list of all Memberships with extended information including email
// @Tags Memberships
// @Accept json
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param limit query int false "Maximum number of results to return"
// @Param offset query int false "Number of results to skip"
// @Success 200 {array} MembershipWithAttendance
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/memberships [get]
func (c *membershipsController) listMemberships(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	filter := c.parseListMembershipsQueryParams(ctx)
	filter.SemesterID = &semesterID

	svc := services.NewMembershipService(c.db)
	memberships, err := svc.ListMembershipsV2(filter)
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

	ctx.JSON(http.StatusOK, memberships)
}

// getMembership handles retrieving a specific membership
//
// @Summary Get a Membership
// @Description Retrieve details of a specific Membership by ID
// @Tags Memberships
// @Accept json
// @Produce json
// @param semesterId path string true "Semester ID"
// @Param id path string true "Membership ID"
// @Success 200 {object} Membership
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/memberships/{id} [get]
func (c *membershipsController) getMembership(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	membershipID, err := validateMembershipID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewMembershipService(c.db)
	membership, err := svc.GetMembershipV2(membershipID, semesterID)
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

	if membership == nil {
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			apierrors.NotFound("Membership with ID '"+membershipID.String()+"' not found"),
		)
		return
	}

	ctx.JSON(http.StatusOK, membership)
}

// updateMembership handles updating a specific membership
//
// @Summary Update a Membership
// @Description Update details of a specific Membership by ID
// @Tags Memberships
// @Accept json
// @Produce json
// @param semesterId path string true "Semester ID"
// @Param id path string true "Membership ID"
// @Param membership body UpdateMembershipRequest true "Updated Membership details"
// @Success 200 {object} Membership
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/memberships/{id} [patch]
func (c *membershipsController) updateMembership(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	membershipID, err := validateMembershipID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	var req models.UpdateMembershipRequestV2
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewMembershipService(c.db)
	membership, err := svc.UpdateMembershipV2(membershipID, semesterID, &req)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		// Check for validation errors
		errMsg := err.Error()
		if errMsg == "cannot set membership to not paid and discounted" {
			ctx.AbortWithStatusJSON(
				http.StatusBadRequest,
				apierrors.InvalidRequest(errMsg),
			)
			return
		}

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	if membership == nil {
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			apierrors.NotFound("Membership with ID '"+membershipID.String()+"' not found"),
		)
		return
	}

	ctx.JSON(http.StatusOK, membership)
}
