package controller

import (
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"net/http"

	apierrors "api/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type semestersController struct {
	db *gorm.DB
}

func NewSemestersController(db *gorm.DB) Controller {
	return &semestersController{db: db}
}

func (s *semestersController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters", middleware.UseAuthentication(s.db))
	group.POST("", middleware.UseAuthorization(s.db, "semester.create"), s.createSemester)
	group.GET("", middleware.UseAuthorization(s.db, "semester.list"), s.listSemesters)
	group.GET(":id", middleware.UseAuthorization(s.db, "semester.get"), s.getSemester)
}

// createSemester handles the creation of a new semester.
// It expects a CreateSemesterRequest in the request body and returns the created Semester.
//
// @Summary Create Semester
// @Description Create a new semester
// @Tags Semesters
// @Accept json
// @Produce json
// @Param semester body CreateSemesterRequest true "Semester data"
// @Success 201 {object} Semester
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters [post]
func (s *semestersController) createSemester(ctx *gin.Context) {
	// Retrieve the request body and bind it to CreateSemesterRequest
	var req models.CreateSemesterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	// Initialize the semester service and create the semester
	svc := services.NewSemesterService(s.db)
	semester, err := svc.CreateSemester(&req)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, semester)
}

// listSemesters handles listing all semesters.
// It returns a list of Semester objects.
//
// @Summary List Semesters
// @Description List all semesters
// @Tags Semesters
// @Produce json
// @Success 200 {array} Semester
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters [get]
func (s *semestersController) listSemesters(ctx *gin.Context) {
	svc := services.NewSemesterService(s.db)
	semesters, err := svc.ListSemesters()
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, semesters)
}

// getSemester handles retrieving a specific semester by ID.
// It returns the Semester object if found.
//
// @Summary Get Semester
// @Description Get a specific semester by ID
// @Tags Semesters
// @Produce json
// @Param id path string true "Semester ID"
// @Success 200 {object} Semester
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{id} [get]
func (s *semestersController) getSemester(c *gin.Context) {
	// Retrieve semester UUID from query parameters
	semesterId := c.Param("id")
	id, err := uuid.Parse(semesterId)
	if err != nil {
		c.JSON(http.StatusBadRequest, apierrors.InvalidRequest("Invalid UUID for semester ID"))
		return
	}

	// Initialize the semester service and get the semester by ID
	svc := services.NewSemesterService(s.db)
	semester, err := svc.GetSemester(id)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			c.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, semester)
}
