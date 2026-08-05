package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresMembershipRepository struct {
	db *gorm.DB
}

var _ store.MembershipRepository = (*postgresMembershipRepository)(nil)

func NewMembershipRepository(db *gorm.DB) store.MembershipRepository {
	return &postgresMembershipRepository{db: db}
}

func (r *postgresMembershipRepository) Create(membership *models.Membership) error {
	return r.db.Create(membership).Error
}

func (r *postgresMembershipRepository) FindByID(id uuid.UUID) (models.Membership, error) {
	var membership models.Membership
	if err := r.db.First(&membership, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Membership{}, store.ErrNotFound
		}

		return models.Membership{}, err
	}

	return membership, nil
}

func (r *postgresMembershipRepository) FindByIDAndSemesterID(id uuid.UUID, semesterID uuid.UUID) (models.Membership, error) {
	var membership models.Membership
	if err := r.db.First(&membership, "id = ? AND semester_id = ?", id, semesterID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Membership{}, store.ErrNotFound
		}

		return models.Membership{}, err
	}

	return membership, nil
}

func (r *postgresMembershipRepository) List(filter *models.ListMembershipsFilter) ([]models.Membership, int64, error) {
	var memberships []models.Membership
	var total int64

	base := r.db.Model(&models.Membership{})

	if filter.SemesterID != nil {
		base = base.Where("semester_id = ?", *filter.SemesterID)
	}
	if filter.UserID != nil {
		base = base.Where("user_id = ?", *filter.UserID)
	}
	if filter.Paid != nil {
		base = base.Where("paid = ?", *filter.Paid)
	}
	if filter.Discounted != nil {
		base = base.Where("discounted = ?", *filter.Discounted)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := base.Order("user_id ASC")
	query = filter.Pagination.Apply(query)

	if err := query.Find(&memberships).Error; err != nil {
		return nil, 0, err
	}

	return memberships, total, nil
}

func (r *postgresMembershipRepository) Update(membership *models.Membership) error {
	return r.db.Model(membership).Select("paid", "discounted").Updates(membership).Error
}

func (r *postgresMembershipRepository) Delete(id uuid.UUID, semesterID uuid.UUID) error {
	result := r.db.Where("semester_id = ?", semesterID).Delete(&models.Membership{}, "id = ?", id)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}
