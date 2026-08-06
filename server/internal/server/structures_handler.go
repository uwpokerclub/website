package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"errors"
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

	blinds := make([]models.Blind, len(req.Blinds))
	for i, blind := range req.Blinds {
		blinds[i] = models.Blind{
			Small: blind.Small,
			Big:   blind.Big,
			Ante:  blind.Ante,
			Time:  blind.Time,
			Index: int8(i),
		}
	}

	structure := models.Structure{
		Name:   req.Name,
		Blinds: blinds,
	}

	if err := s.store.Structures().Create(&structure); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, structure)
}

func (s *apiServer) ListStructures(ctx *gin.Context) {
	structures, _, err := s.store.Structures().List(&models.Pagination{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
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

	structure, err := s.store.Structures().FindByID(int32(structureId))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
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
	req.ID = int32(structureId)

	structure, err := s.store.Structures().FindByID(req.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, e.NotFound(err.Error()))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}
	structure.Name = req.Name

	blinds := make([]models.Blind, len(req.Blinds))
	for i, blind := range req.Blinds {
		blinds[i] = models.Blind{
			Small:       blind.Small,
			Big:         blind.Big,
			Ante:        blind.Ante,
			Time:        blind.Time,
			StructureId: req.ID,
			Index:       int8(i),
		}
	}

	tx, err := s.store.BeginTx()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}
	defer tx.Rollback()

	if err := tx.Structures().Update(&structure); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	if err := tx.Structures().ReplaceBlindsByStructureID(req.ID, blinds); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	if err := tx.Commit(); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	result, err := s.store.Structures().FindByID(req.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, result)
}
