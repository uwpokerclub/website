package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rankingsController struct {
	db *gorm.DB
}

// NewRankingsController creates a new instance of rankingsController
func NewRankingsController(db *gorm.DB) Controller {
	return &rankingsController{db: db}
}

func (c *rankingsController) LoadRoutes(router *gin.RouterGroup) {
	rankings := router.Group("semesters/:semesterId/rankings", middleware.UseAuthentication(c.db))
	rankings.GET("", middleware.UseAuthorization(c.db, "semester.rankings.list"), c.listRankings)
	rankings.GET("export", middleware.UseAuthorization(c.db, "semester.rankings.export"), c.exportRankings)
	rankings.GET(":membershipId", middleware.UseAuthorization(c.db, "semester.rankings.get"), c.getRanking)
}

func validateUUIDParam(ctx *gin.Context, paramName string) (uuid.UUID, error) {
	idParam := ctx.Param(paramName)
	id, err := uuid.Parse(idParam)
	if err != nil {
		return uuid.Nil, errors.New(
			paramName + " '" + idParam + "' is not a valid UUID",
		)
	}

	return id, nil
}

// listRankings handles listing the current rankings for a semester
//
// @Summary List rankings
// @Description List the current rankings for a semester
// @Tags Rankings
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Success 200 {array} RankingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/rankings [get]
func (c *rankingsController) listRankings(ctx *gin.Context) {
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

	svc := services.NewSemesterService(c.db)
	rankings, total, err := svc.GetRankingsV2(semesterID, &pagination)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.ListResponse[models.RankingResponse]{
		Data:  rankings,
		Total: total,
	})
}

// getRanking handles retrieving the ranking for a specific membership in a semester
//
// @Summary Get ranking
// @Description Get the ranking for a specific membership in a semester
// @Tags Rankings
// @Produce json
// @Param semesterId path string true "Semester ID"
// @Param membershipId path string true "Membership ID"
// @Success 200 {object} GetRankingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/rankings/{membershipId} [get]
func (c *rankingsController) getRanking(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	membershipID, err := validateUUIDParam(ctx, "membershipId")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewRankingService(c.db)
	ranking, err := svc.GetRanking(semesterID, membershipID)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, ranking)
}

// exportRankings handles exporting the rankings for a semester
//
// @Summary Export rankings
// @Description Export the rankings for a semester
// @Tags Rankings
// @Produce octet-stream
// @Param semesterId path string true "Semester ID"
// @Success 200 {file} file "Rankings exported successfully"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /semesters/{semesterId}/rankings/export [get]
func (c *rankingsController) exportRankings(ctx *gin.Context) {
	semesterID, err := validateSemesterID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewSemesterService(c.db)
	fp, err := svc.ExportRankings(semesterID)
	if err != nil {
		if apiErr, ok := err.(apierrors.APIErrorResponse); ok {
			ctx.AbortWithStatusJSON(apiErr.Code, apiErr)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.InternalServerError(err.Error()))
		return
	}

	filename := filepath.Base(fp)
	defer func() {
		err := os.Remove(fp)
		if err != nil {
			// Log the error but do not interrupt the response
			log.Println("Failed to remove temporary file:", err)
		}
	}()

	ctx.FileAttachment(fp, filename)
}
