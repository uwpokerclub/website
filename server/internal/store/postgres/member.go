package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"gorm.io/gorm"
)

type postgresMemberRepository struct {
	db *gorm.DB
}

var _ store.MemberRepository = (*postgresMemberRepository)(nil)

func NewMemberRepository(db *gorm.DB) store.MemberRepository {
	return &postgresMemberRepository{db: db}
}

func (r *postgresMemberRepository) Create(member *models.User) error {
	return r.db.Create(member).Error
}

func (r *postgresMemberRepository) FindByID(id uint64) (models.User, error) {
	var member models.User
	if err := r.db.First(&member, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, store.ErrNotFound
		}

		return models.User{}, err
	}

	return member, nil
}

func (r *postgresMemberRepository) List(filter *models.ListUsersFilter, pagination *models.Pagination) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	base := r.db.Model(&models.User{})

	if filter.ID != nil {
		base = base.Where("id = ?", *filter.ID)
	}
	if filter.Name != nil {
		base = base.Where("first_name || ' ' || last_name ILIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Email != nil {
		base = base.Where("email ILIKE ?", "%"+*filter.Email+"%")
	}
	if filter.Faculty != nil {
		base = base.Where("faculty = ?", *filter.Faculty)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := base.Order("created_at DESC")
	query = pagination.Apply(query)

	if err := query.Find(&members).Error; err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

func (r *postgresMemberRepository) Update(member *models.User) error {
	return r.db.Model(member).
		Select("first_name", "last_name", "email", "faculty", "quest_id").
		Updates(member).Error
}

func (r *postgresMemberRepository) Delete(id uint64) error {
	result := r.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}
