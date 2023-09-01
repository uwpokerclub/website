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

func (ss *structureService) CreateStructure(req *models.CreateStructureRequest) (*models.StructureWithBlindsResponse, error) {
	structure := models.Structure{
		Name: req.Name,
	}

	tx := ss.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	res := tx.Create(&structure)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	blinds := make([]models.Blind, len(req.Blinds))
	for i, blind := range req.Blinds {
		blinds[i] = models.Blind{
			Small:       blind.Small,
			Big:         blind.Big,
			Ante:        blind.Ante,
			Time:        blind.Time,
			StructureId: structure.ID,
			Index:       i,
		}
	}

	res = tx.Create(&blinds)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
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

	return &models.StructureWithBlindsResponse{
		ID:     structure.ID,
		Name:   structure.Name,
		Blinds: blindsRes,
	}, nil
}

func (ss *structureService) ListStructures() ([]models.Structure, error) {
	var structures []models.Structure

	res := ss.db.Order("id DESC").Find(&structures)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return structures, nil
}

func (ss *structureService) GetStructure(id uint64) (*models.StructureWithBlindsResponse, error) {
	structure := models.Structure{
		ID: id,
	}

	res := ss.db.First(&structure)
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

	return &models.StructureWithBlindsResponse{
		ID:     structure.ID,
		Name:   structure.Name,
		Blinds: blindsRes,
	}, nil
}

func (ss *structureService) UpdateStructure(req *models.UpdateStructureRequest) (*models.StructureWithBlindsResponse, error) {
	structure := models.Structure{
		ID: req.ID,
	}

	res := ss.db.First(&structure)
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	} else if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	tx := ss.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	structure.Name = req.Name
	res = tx.Save(&structure)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	// Delete old blind levels
	res = tx.Where("structure_id = ?", structure.ID).Delete(&models.Blind{})
	if err := res.Error; err != nil {
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
			Index:       i,
		}
	}

	res = tx.Create(&blinds)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
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

	return &models.StructureWithBlindsResponse{
		ID:     structure.ID,
		Name:   structure.Name,
		Blinds: blindsRes,
	}, nil
}
