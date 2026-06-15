package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"gorm.io/gorm"
)

type postgresStructureRepository struct {
	db *gorm.DB
}

var _ store.StructureRepository = (*postgresStructureRepository)(nil)

func NewStructureRepository(db *gorm.DB) store.StructureRepository {
	return &postgresStructureRepository{db: db}
}

func (r *postgresStructureRepository) Create(structure *models.Structure) error {
	return r.db.Create(structure).Error
}

func (r *postgresStructureRepository) FindByID(id int32) (models.Structure, error) {
	structure := models.Structure{
		ID: id,
	}

	res := structure.Preload(r.db, models.StructurePreloadOptions{Blinds: true}).First(&structure)	
	if err := res.Error; err != nil {	
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Structure{}, store.ErrNotFound
		}

		return models.Structure{}, err
	}

	return structure, nil
}

func (r *postgresStructureRepository) List(pagination *models.Pagination) ([]models.Structure, int64, error) {
	var structures []models.Structure
	var total int64

	base := r.db.Model(&models.Structure{})

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := base.Order("id DESC")
	query = pagination.Apply(query)

	if err := query.Find(&structures).Error; err != nil {
		return nil, 0, err
	}

	return structures, total, nil
}

func (r *postgresStructureRepository) Update(structure *models.Structure) error {
	return r.db.Model(structure).Select("name").Updates(structure).Error
}

func (r *postgresStructureRepository) ReplaceBlindsByStructureID(structureID int32, blinds []models.Blind) error {
	if err := r.db.Where("structure_id = ?", structureID).Delete(&models.Blind{}).Error; err != nil {
		return err
	}
	if len(blinds) > 0 {
		return r.db.Create(&blinds).Error
	}
	// FK constraint only catches non-existent structure when blinds are inserted.
	// With an empty slice, verify the structure exists explicitly.
	var count int64
	if err := r.db.Model(&models.Structure{}).Where("id = ?", structureID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *postgresStructureRepository) Delete(id int32) error {
	result := r.db.Delete(&models.Structure{}, id)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}
