package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *apiServer) CreateSemester(ctx *gin.Context) {
	var req models.CreateSemesterRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewSemesterService(s.db)
	semester, err := svc.CreateSemester(&req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, semester)

}

func (s *apiServer) ListSemesters(ctx *gin.Context) {
	svc := services.NewSemesterService(s.db)
	semesters, err := svc.ListSemesters()
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
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

	svc := services.NewSemesterService(s.db)
	semester, err := svc.GetSemester(id)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
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

	svc := services.NewSemesterService(s.db)
	rankings, err := svc.GetRankings(id)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, rankings)
}
