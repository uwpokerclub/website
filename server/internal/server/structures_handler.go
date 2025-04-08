package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *apiServer) CreateStructure(ctx *gin.Context) {
	var req models.CreateStructureRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewStructureService(s.db)
	structure, err := svc.CreateStructure(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, structure)
}

func (s *apiServer) ListStructures(ctx *gin.Context) {
	svc := services.NewStructureService(s.db)
	structures, err := svc.ListStructures()
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, structures)
}

func (s *apiServer) GetStructure(ctx *gin.Context) {
	structureId, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid structure ID specified in request"))
		return
	}

	svc := services.NewStructureService(s.db)
	structure, err := svc.GetStructure(structureId)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, structure)
}

func (s *apiServer) UpdateStructure(ctx *gin.Context) {
	structureId, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest("Invalid structure ID specified in request"))
		return
	}

	var req models.UpdateStructureRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
		return
	}
	req.ID = structureId

	svc := services.NewStructureService(s.db)
	structure, err := svc.UpdateStructure(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, structure)
}
