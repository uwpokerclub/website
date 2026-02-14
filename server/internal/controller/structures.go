package controller

import (
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type structuresController struct {
	db *gorm.DB
}

func NewStructuresController(db *gorm.DB) Controller {
	return &structuresController{db: db}
}

func (s *structuresController) LoadRoutes(router *gin.RouterGroup) {
	group := router.Group("structures", middleware.UseAuthentication(s.db))
	group.GET("", middleware.UseAuthorization("structure.list"), s.listStructures)
	group.POST("", middleware.UseAuthorization("structure.create"), s.createStructure)
	group.GET(":id", middleware.UseAuthorization("structure.get"), s.getStructure)
	group.PATCH(":id", middleware.UseAuthorization("structure.edit"), s.updateStructure)
	group.DELETE(":id", middleware.UseAuthorization("structure.delete"), s.deleteStructure)
}

// listStructures handles the retrieval of all structures.
// It returns a list of structures without their blinds.
//
// @Summary List Structures
// @Description Retrieve a list of all structures
// @Tags Structures
// @Accept json
// @Produce json
// @Success 200 {array} models.Structure
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /structures [get]
func (s *structuresController) listStructures(ctx *gin.Context) {
	pagination, err := models.ParsePagination(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewStructureService(s.db)
	structures, total, err := svc.ListStructuresV2(&pagination)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusOK, models.ListResponse[models.Structure]{
		Data:  structures,
		Total: total,
	})
}

// createStructure handles the creation of a new structure.
// It expects a CreateStructureRequest in the request body and returns the created Structure.
//
// @Summary Create Structure
// @Description Create a new structure with blind levels
// @Tags Structures
// @Accept json
// @Produce json
// @Param structure body models.CreateStructureRequest true "Structure data"
// @Success 201 {object} models.Structure
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /structures [post]
func (s *structuresController) createStructure(ctx *gin.Context) {
	var req models.CreateStructureRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewStructureService(s.db)
	structure, err := svc.CreateStructure(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			apierrors.InternalServerError(err.Error()),
		)
		return
	}

	ctx.JSON(http.StatusCreated, structure)
}

// getStructure handles the retrieval of a specific structure by its ID.
// It expects the structure ID as a URL parameter.
//
// @Summary Get Structure
// @Description Retrieve a specific structure by its ID with blinds
// @Tags Structures
// @Accept json
// @Produce json
// @Param id path string true "Structure ID"
// @Success 200 {object} models.Structure
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /structures/{id} [get]
func (s *structuresController) getStructure(ctx *gin.Context) {
	id, err := s.parseStructureID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewStructureService(s.db)
	structure, err := svc.GetStructure(id)
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

	ctx.JSON(http.StatusOK, structure)
}

// updateStructure handles the partial update of an existing structure.
// It expects the structure ID as a URL parameter and partial update data in the body.
//
// @Summary Update Structure
// @Description Partially update an existing structure
// @Tags Structures
// @Accept json
// @Produce json
// @Param id path string true "Structure ID"
// @Param structure body object true "Partial structure data"
// @Success 200 {object} models.Structure
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /structures/{id} [patch]
func (s *structuresController) updateStructure(ctx *gin.Context) {
	id, err := s.parseStructureID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	requestValues := make(map[string]any)
	if err := ctx.ShouldBindBodyWithJSON(&requestValues); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(
				fmt.Sprintf("Error parsing request body: %s", err.Error()),
			),
		)
		return
	}

	updateMap, err := s.validateAndCreateStructureUpdateMap(requestValues)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			apierrors.InvalidRequest(err.Error()),
		)
		return
	}

	svc := services.NewStructureService(s.db)
	structure, err := svc.UpdateStructureV2(id, updateMap)
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

	ctx.JSON(http.StatusOK, structure)
}

// deleteStructure handles the deletion of an existing structure.
// It expects the structure ID as a URL parameter.
//
// @Summary Delete Structure
// @Description Delete an existing structure
// @Tags Structures
// @Accept json
// @Produce json
// @Param id path string true "Structure ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /structures/{id} [delete]
func (s *structuresController) deleteStructure(ctx *gin.Context) {
	id, err := s.parseStructureID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apierrors.InvalidRequest(err.Error()))
		return
	}

	svc := services.NewStructureService(s.db)
	err = svc.DeleteStructure(id)
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

// parseStructureID parses and validates the structure ID from the URL parameter
func (s *structuresController) parseStructureID(ctx *gin.Context) (int32, error) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Structure ID '%s' is not a valid integer", idParam)
	}
	if id <= 0 {
		return 0, errors.New("Structure ID must be a positive integer")
	}
	return int32(id), nil
}

// validateAndCreateStructureUpdateMap validates the request values and creates an update map
func (s *structuresController) validateAndCreateStructureUpdateMap(
	requestValues map[string]any,
) (map[string]any, error) {
	updateMap := make(map[string]any)

	for key, value := range requestValues {
		switch key {
		case "name":
			if value == nil {
				return nil, errors.New("name cannot be null")
			}
			strValue, ok := value.(string)
			if !ok {
				return nil, errors.New("name must be a string")
			}
			if strValue == "" {
				return nil, errors.New("name must be a non-empty string")
			}
			updateMap["name"] = strValue
		case "blinds":
			if value == nil {
				return nil, errors.New("blinds cannot be null")
			}
			blindsArray, ok := value.([]any)
			if !ok {
				return nil, errors.New("blinds must be an array")
			}
			if len(blindsArray) == 0 {
				return nil, errors.New("blinds array cannot be empty")
			}
			blinds := make([]models.BlindJSON, len(blindsArray))
			for i, b := range blindsArray {
				blindMap, ok := b.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("blind at index %d must be an object", i)
				}
				blind, err := s.parseBlindJSON(blindMap, i)
				if err != nil {
					return nil, err
				}
				blinds[i] = blind
			}
			updateMap["blinds"] = blinds
		default:
			return nil, fmt.Errorf("unknown field: %s", key)
		}
	}

	return updateMap, nil
}

// parseBlindJSON parses and validates a blind JSON object
func (s *structuresController) parseBlindJSON(blindMap map[string]any, index int) (models.BlindJSON, error) {
	blind := models.BlindJSON{}

	// Parse small
	smallVal, ok := blindMap["small"]
	if !ok {
		return blind, fmt.Errorf("blind at index %d is missing 'small' field", index)
	}
	smallFloat, ok := smallVal.(float64)
	if !ok {
		return blind, fmt.Errorf("blind at index %d: 'small' must be a number", index)
	}
	if smallFloat < 0 {
		return blind, fmt.Errorf("blind at index %d: 'small' must be >= 0", index)
	}
	blind.Small = int32(smallFloat)

	// Parse big
	bigVal, ok := blindMap["big"]
	if !ok {
		return blind, fmt.Errorf("blind at index %d is missing 'big' field", index)
	}
	bigFloat, ok := bigVal.(float64)
	if !ok {
		return blind, fmt.Errorf("blind at index %d: 'big' must be a number", index)
	}
	if bigFloat < 0 {
		return blind, fmt.Errorf("blind at index %d: 'big' must be >= 0", index)
	}
	blind.Big = int32(bigFloat)

	// Parse ante (optional, defaults to 0)
	if anteVal, ok := blindMap["ante"]; ok {
		anteFloat, ok := anteVal.(float64)
		if !ok {
			return blind, fmt.Errorf("blind at index %d: 'ante' must be a number", index)
		}
		if anteFloat < 0 {
			return blind, fmt.Errorf("blind at index %d: 'ante' must be >= 0", index)
		}
		blind.Ante = int32(anteFloat)
	}

	// Parse time
	timeVal, ok := blindMap["time"]
	if !ok {
		return blind, fmt.Errorf("blind at index %d is missing 'time' field", index)
	}
	timeFloat, ok := timeVal.(float64)
	if !ok {
		return blind, fmt.Errorf("blind at index %d: 'time' must be a number", index)
	}
	if timeFloat <= 0 || timeFloat > 60 {
		return blind, fmt.Errorf("blind at index %d: 'time' must be between 1 and 60", index)
	}
	blind.Time = int8(timeFloat)

	return blind, nil
}
