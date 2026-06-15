package controller

import (
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/store"
	"errors"
	"net/http"

	apierrors "api/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type semestersController struct {
	db    *gorm.DB
	store store.Store
}

func NewSemestersController(db *gorm.DB, st store.Store) Controller {
	return &semestersController{db: db, store: st}
}

func (s *semestersController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("semesters", middleware.UseAuthentication(s.db))
	group.POST("", middleware.UseAuthorization("semester.create"), s.createSemester)
	group.GET("", middleware.UseAuthorization("semester.list"), s.listSemesters)
	group.GET(":semesterId", middleware.UseAuthorization("semester.get"), s.getSemester)
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
	var req models.CreateSemesterRequest
	if !BindJSON(ctx, &req) {
		return
	}

	semester := models.Semester{
		Name:                  req.Name,
		Meta:                  req.Meta,
		StartDate:             req.StartDate,
		EndDate:               req.EndDate,
		StartingBudget:        req.StartingBudget,
		CurrentBudget:         req.StartingBudget,
		MembershipFee:         req.MembershipFee,
		MembershipDiscountFee: req.MembershipDiscountFee,
		RebuyFee:              req.RebuyFee,
	}

	if err := s.store.Semesters().Create(&semester); err != nil {
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
	pagination, err := models.ParsePagination(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	semesters, total, err := s.store.Semesters().List(&pagination)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.ListResponse[models.Semester]{
		Data:  semesters,
		Total: total,
	})
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
	semesterId := c.Param("semesterId")
	id, err := uuid.Parse(semesterId)
	if err != nil {
		c.JSON(http.StatusBadRequest, apierrors.InvalidRequest("Invalid UUID for semester ID"))
		return
	}

	semester, err := s.store.Semesters().FindByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, apierrors.NotFound(err.Error()))
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, semester)
}
