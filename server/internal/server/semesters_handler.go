package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"api/internal/store"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func (s *apiServer) CreateSemester(ctx *gin.Context) {
	var req models.CreateSemesterRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
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
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, semester)
}

func (s *apiServer) ListSemesters(ctx *gin.Context) {
	semesters, _, err := s.store.Semesters().List(&models.Pagination{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, semesters)
}

func (s *apiServer) GetSemester(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	id, err := uuid.Parse(semesterId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
		return
	}

	semester, err := s.store.Semesters().FindByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, semester)
}

func (s *apiServer) GetRankings(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	id, err := uuid.Parse(semesterId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
		return
	}

	rankings, _, err := s.store.Rankings().List(&models.ListRankingsFilter{SemesterID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, rankings)
}

func (s *apiServer) GetRanking(ctx *gin.Context) {
	queryValue := ctx.Param("semesterId")
	semesterID, err := uuid.Parse(queryValue)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
		return
	}

	queryValue = ctx.Param("membershipId")
	membershipID, err := uuid.Parse(queryValue)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for membership ID"))
		return
	}

	ranking, err := s.store.Rankings().FindBySemesterAndMembershipID(semesterID, membershipID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, ranking)
}

func (s *apiServer) ExportRankings(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	id, err := uuid.Parse(semesterId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid semester ID specified in request"))
		return
	}

	svc := services.NewSemesterService(s.store)

	fp, err := svc.ExportRankings(id)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	filename := filepath.Base(fp)
	defer os.Remove(filename)

	ctx.FileAttachment(fp, filename)
}
