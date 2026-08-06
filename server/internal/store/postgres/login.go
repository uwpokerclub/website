package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"gorm.io/gorm"
)

type postgresLoginRepository struct {
	db *gorm.DB
}

var _ store.LoginRepository = (*postgresLoginRepository)(nil)

func NewLoginRepository(db *gorm.DB) store.LoginRepository {
	return &postgresLoginRepository{db: db}
}

func (r *postgresLoginRepository) Create(login *models.Login) error {
	return r.db.Create(login).Error
}

func (r *postgresLoginRepository) FindByUsername(username string) (models.Login, error) {
	var login models.Login

	err := r.db.Where("username = ?", username).First(&login).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Login{}, store.ErrNotFound
		}
		return models.Login{}, err
	}

	return login, nil
}

func (r *postgresLoginRepository) Update(username string, values map[string]any) error {
	result := r.db.Model(&models.Login{}).Where("username = ?", username).Updates(values)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *postgresLoginRepository) Delete(username string) error {
	result := r.db.Where("username = ?", username).Delete(&models.Login{})
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}
