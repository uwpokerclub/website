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
			Index: i,
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

func (ss *structureService) GetStructure(id uint64) (*models.Structure, error) {
	structure := models.Structure{
		ID: id,
	}

	res := ss.db.Model(&structure).Preload("Blinds").First(&structure)
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

	res := ss.db.Model(&structure).Preload("Blinds").First(&structure)

	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	} else if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	structure.Name = req.Name

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
	err := ss.db.Model(&structure).Association("Blinds").Replace(blinds)
	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	ss.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&structure)

	return &structure, nil
}
