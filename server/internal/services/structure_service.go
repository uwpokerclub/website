package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"

	"gorm.io/gorm"
)

type structureService struct {
	db *gorm.DB
}

func NewStructureService(db *gorm.DB) *structureService {
	return &structureService{
		db: db,
	}
}

func (ss *structureService) CreateStructure(req *models.CreateStructureRequest) (*models.Structure, error) {
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

	res := ss.db.Create(&structure)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	blindsRes := make([]models.BlindJSON, len(blinds))
	for i, blind := range blinds {
		blindsRes[i] = models.BlindJSON{
			Small: blind.Small,
			Big:   blind.Big,
			Ante:  blind.Ante,
			Time:  blind.Time,
		}
	}

	return &structure, nil
}

func (ss *structureService) ListStructures() ([]models.Structure, error) {
	var structures []models.Structure

	res := ss.db.Order("id DESC").Find(&structures)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return structures, nil
}

func (ss *structureService) GetStructure(id int32) (*models.Structure, error) {
	structure := models.Structure{
		ID: id,
	}

	res := structure.Preload(ss.db, models.StructurePreloadOptions{Blinds: true}).First(&structure)
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	} else if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	var blinds []models.Blind
	res = ss.db.Where("structure_id = ?", id).Order("index ASC").Find(&blinds)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	blindsRes := make([]models.BlindJSON, len(blinds))
	for i, blind := range blinds {
		blindsRes[i] = models.BlindJSON{
			Small: blind.Small,
			Big:   blind.Big,
			Ante:  blind.Ante,
			Time:  blind.Time,
		}
	}

	return &structure, nil
}

func (ss *structureService) UpdateStructure(req *models.UpdateStructureRequest) (*models.Structure, error) {
	structure := models.Structure{
		ID: req.ID,
	}

	res := structure.Preload(ss.db, models.StructurePreloadOptions{Blinds: true}).First(&structure)

	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	} else if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Use a transaction to ensure atomicity
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	structure.Name = req.Name

	// Update the structure first
	err := tx.Updates(&structure).Error
	if err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	// Delete existing blinds
	err = tx.Where("structure_id = ?", structure.ID).Delete(&models.Blind{}).Error
	if err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	// Insert new levels
	blinds := make([]models.Blind, len(req.Blinds))
	for i, blind := range req.Blinds {
		blinds[i] = models.Blind{
			Small:       blind.Small,
			Big:         blind.Big,
			Ante:        blind.Ante,
			Time:        blind.Time,
			StructureId: structure.ID,
			Index:       int8(i),
		}
	}

	// Create new blinds
	if len(blinds) > 0 {
		err = tx.Create(&blinds).Error
		if err != nil {
			tx.Rollback()
			return nil, e.InternalServerError(err.Error())
		}
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Reload with blinds using a fresh query
	result := models.Structure{}
	err = result.Preload(ss.db, models.StructurePreloadOptions{Blinds: true}).Where("id = ?", structure.ID).First(&result).Error
	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &result, nil
}

// UpdateStructureV2 performs a partial update - only updates fields that are provided in updateMap
func (ss *structureService) UpdateStructureV2(id int32, updateMap map[string]any) (*models.Structure, error) {
	structure := models.Structure{
		ID: id,
	}

	res := structure.Preload(ss.db, models.StructurePreloadOptions{Blinds: true}).First(&structure)
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound("Structure not found")
	} else if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// If nothing to update, return existing structure
	if len(updateMap) == 0 {
		return &structure, nil
	}

	// Use a transaction to ensure atomicity
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update name if provided
	if name, ok := updateMap["name"]; ok {
		structure.Name = name.(string)
		err := tx.Model(&structure).Update("name", name).Error
		if err != nil {
			tx.Rollback()
			return nil, e.InternalServerError(err.Error())
		}
	}

	// Update blinds if provided (full replacement)
	if blindsData, ok := updateMap["blinds"]; ok {
		// Delete existing blinds
		err := tx.Where("structure_id = ?", structure.ID).Delete(&models.Blind{}).Error
		if err != nil {
			tx.Rollback()
			return nil, e.InternalServerError(err.Error())
		}

		// Insert new blinds
		blindsJSON := blindsData.([]models.BlindJSON)
		blinds := make([]models.Blind, len(blindsJSON))
		for i, b := range blindsJSON {
			blinds[i] = models.Blind{
				Small:       b.Small,
				Big:         b.Big,
				Ante:        b.Ante,
				Time:        b.Time,
				StructureId: structure.ID,
				Index:       int8(i),
			}
		}

		if len(blinds) > 0 {
			err = tx.Create(&blinds).Error
			if err != nil {
				tx.Rollback()
				return nil, e.InternalServerError(err.Error())
			}
		}
	}

	// Commit the transaction
	err := tx.Commit().Error
	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Reload with blinds using a fresh query
	result := models.Structure{}
	err = result.Preload(ss.db, models.StructurePreloadOptions{Blinds: true}).Where("id = ?", structure.ID).First(&result).Error
	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &result, nil
}

// DeleteStructure deletes a structure by ID
// Manually deletes associated blinds first since FK constraint is ON DELETE NO ACTION
func (ss *structureService) DeleteStructure(id int32) error {
	structure := models.Structure{
		ID: id,
	}

	res := ss.db.First(&structure)
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return e.NotFound("Structure not found")
	} else if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	// Use transaction to ensure atomicity
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete associated blinds first
	if err := tx.Where("structure_id = ?", id).Delete(&models.Blind{}).Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	// Delete the structure
	if err := tx.Delete(&structure).Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}
