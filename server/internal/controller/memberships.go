package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type membershipsController struct {
	store store.Store
}

// NewMembershipsController creates a new instance of membershipsController
func NewMembershipsController(st store.Store) Controller {
	return &membershipsController{store: st}
}

func (c *membershipsController) LoadRoutes(router *gin.RouterGroup) {
	memberships := router.Group("semesters/:semesterId/memberships", middleware.UseAuthentication(c.store))
	memberships.POST("", middleware.UseAuthorization("membership.create"), c.createMembership)
	memberships.GET("", middleware.UseAuthorization("membership.list"), c.listMemberships)
	memberships.GET("/:id", middleware.UseAuthorization("membership.get"), c.getMembership)
	memberships.PATCH(
		"/:id",
		middleware.UseAuthorization("membership.edit"),
		c.updateMembership,
	)
	memberships.DELETE("/:id", middleware.UseAuthorization("membership.delete"), c.deleteMembership)
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

	var req models.CreateMembershipRequest
	if !BindJSON(ctx, &req) {
		return
	}

	svc := services.NewMembershipService(c.store)
	membership, err := svc.CreateMembership(semesterID, &req)
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
// @Param search query string false "Search by first name, last name, email, or full name"
// @Param name query string false "Filter by first name, last name, or full name (case-insensitive partial match)"
// @Param email query string false "Filter by email (case-insensitive partial match)"
// @Param faculty query string false "Filter by faculty (exact match)" Enums(AHS, Arts, Engineering, Environment, Math, Science)
// @Param studentId query string false "Filter by student ID (exact match)"
// @Param paid query bool false "Filter by paid status"
// @Param discounted query bool false "Filter by discounted status"
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

	pagination, err := models.ParsePagination(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	search := ctx.Query("search")

	filter := &models.ListMembershipsFilter{
		Pagination: pagination,
		SemesterID: &semesterID,
		Search:     search,
	}

	if name := ctx.Query("name"); name != "" {
		filter.Name = &name
	}
	if email := ctx.Query("email"); email != "" {
		filter.Email = &email
	}
	if faculty := ctx.Query("faculty"); faculty != "" {
		filter.Faculty = &faculty
	}
	if studentID := ctx.Query("studentId"); studentID != "" {
		filter.StudentID = &studentID
	}
	if paid := ctx.Query("paid"); paid == "true" || paid == "false" {
		v := paid == "true"
		filter.Paid = &v
	}
	if discounted := ctx.Query("discounted"); discounted == "true" || discounted == "false" {
		v := discounted == "true"
		filter.Discounted = &v
	}

	memberships, total, err := c.store.Memberships().List(filter)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, models.ListResponse[models.MembershipWithAttendance]{
		Data:  memberships,
		Total: total,
	})
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

	membership, err := c.store.Memberships().FindByIDAndSemesterID(membershipID, semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(
				http.StatusNotFound,
				apierrors.NotFound("Membership with ID '"+membershipID.String()+"' not found"),
			)
			return
		}
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
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

	var req models.UpdateMembershipRequest
	if !BindJSON(ctx, &req) {
		return
	}

	svc := services.NewMembershipService(c.store)
	membership, err := svc.UpdateMembership(membershipID, semesterID, &req)
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

// deleteMembership handles deleting a specific membership
//
// @Summary Delete a Membership
// @Description Delete a specific Membership by ID. Rankings are cascade deleted and participant entries are unlinked.
// @Tags Memberships
// @Accept json
// @Produce json
// @param semesterId path string true "Semester ID"
// @Param id path string true "Membership ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/memberships/{id} [delete]
func (c *membershipsController) deleteMembership(ctx *gin.Context) {
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

	err = c.store.Memberships().Delete(membershipID, semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(
				http.StatusNotFound,
				apierrors.NotFound("Membership with ID '"+membershipID.String()+"' not found"),
			)
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
