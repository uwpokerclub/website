package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresSemesterRepository struct {
	db *gorm.DB
}

var _ store.SemesterRepository = (*postgresSemesterRepository)(nil)

func NewSemesterRepository(db *gorm.DB) store.SemesterRepository {
	return &postgresSemesterRepository{db: db}
}

func (r *postgresSemesterRepository) Create(semester *models.Semester) error {
	return r.db.Create(semester).Error
}

func (r *postgresSemesterRepository) FindByID(id uuid.UUID) (*models.Semester, error) {
	var semester models.Semester
	if err := r.db.First(&semester, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &semester, nil
}

func (r *postgresSemesterRepository) List(pagination *models.Pagination) ([]*models.Semester, int64, error) {
	var semesters []*models.Semester
	var total int64

	if err := r.db.Model(&models.Semester{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Order("start_date DESC")
	query = pagination.Apply(query)

	if err := query.Find(&semesters).Error; err != nil {
		return nil, 0, err
	}

	return semesters, total, nil
}

func (r *postgresSemesterRepository) Update(semester *models.Semester) error {
	return r.db.Save(semester).Error
}
