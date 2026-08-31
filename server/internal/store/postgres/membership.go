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
	if err := (models.Membership{}).Preload(r.db).First(&membership, "memberships.id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Membership{}, store.ErrNotFound
		}

		return models.Membership{}, err
	}

	return membership, nil
}

func (r *postgresMembershipRepository) FindByIDAndSemesterID(id uuid.UUID, semesterID uuid.UUID) (models.Membership, error) {
	var membership models.Membership
	if err := (models.Membership{}).Preload(r.db).First(&membership, "memberships.id = ? AND memberships.semester_id = ?", id, semesterID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Membership{}, store.ErrNotFound
		}

		return models.Membership{}, err
	}

	return membership, nil
}

func (r *postgresMembershipRepository) applyJoinedFilterClauses(query *gorm.DB, filter *models.ListMembershipsFilter) *gorm.DB {
	if filter.Search != "" {
		pattern := "%" + eventNameLikeReplacer.Replace(filter.Search) + "%"
		query = query.Where(
			`"User".first_name ILIKE ? OR "User".last_name ILIKE ? OR "User".email ILIKE ? OR ("User".first_name || ' ' || "User".last_name) ILIKE ?`,
			pattern, pattern, pattern, pattern,
		)
	}
	if filter.Name != nil {
		pattern := "%" + eventNameLikeReplacer.Replace(*filter.Name) + "%"
		query = query.Where(
			`"User".first_name ILIKE ? OR "User".last_name ILIKE ? OR ("User".first_name || ' ' || "User".last_name) ILIKE ?`,
			pattern, pattern, pattern,
		)
	}
	if filter.Email != nil {
		pattern := "%" + eventNameLikeReplacer.Replace(*filter.Email) + "%"
		query = query.Where(`"User".email ILIKE ?`, pattern)
	}
	if filter.Faculty != nil {
		query = query.Where(`"User".faculty = ?`, *filter.Faculty)
	}
	if filter.StudentID != nil {
		query = query.Where(`CAST("User".id AS TEXT) = ?`, *filter.StudentID)
	}
	return query
}

func (r *postgresMembershipRepository) List(filter *models.ListMembershipsFilter) ([]models.MembershipWithAttendance, int64, error) {
	needsUserJoin := filter.Search != "" || filter.Name != nil || filter.Email != nil || filter.Faculty != nil || filter.StudentID != nil

	countQuery := r.db.Model(&models.Membership{})
	if filter.SemesterID != nil {
		countQuery = countQuery.Where("memberships.semester_id = ?", *filter.SemesterID)
	}
	if filter.UserID != nil {
		countQuery = countQuery.Where("memberships.user_id = ?", *filter.UserID)
	}
	if filter.Paid != nil {
		countQuery = countQuery.Where("memberships.paid = ?", *filter.Paid)
	}
	if filter.Discounted != nil {
		countQuery = countQuery.Where("memberships.discounted = ?", *filter.Discounted)
	}
	if filter.Executive != nil {
		countQuery = countQuery.Where("memberships.executive = ?", *filter.Executive)
	}
	if needsUserJoin {
		countQuery = countQuery.Joins("User")
	}
	countQuery = r.applyJoinedFilterClauses(countQuery, filter)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	attendanceSubquery := r.db.
		Select("participants.membership_id, COUNT(*) as total").
		Table("participants").
		Joins("INNER JOIN events ON participants.event_id = events.id")
	if filter.SemesterID != nil {
		attendanceSubquery = attendanceSubquery.Where("events.semester_id = ?", *filter.SemesterID)
	}
	attendanceSubquery = attendanceSubquery.Group("participants.membership_id")

	query := r.db.
		Select("memberships.*, COALESCE(att.total, 0) as attendance").
		Joins("User").
		Joins("LEFT JOIN (?) as att ON att.membership_id = memberships.id", attendanceSubquery).
		Order(`"User".first_name ASC`).
		Order(`"User".last_name ASC`)
	if filter.SemesterID != nil {
		query = query.Where("memberships.semester_id = ?", *filter.SemesterID)
	}
	if filter.UserID != nil {
		query = query.Where("memberships.user_id = ?", *filter.UserID)
	}
	if filter.Paid != nil {
		query = query.Where("memberships.paid = ?", *filter.Paid)
	}
	if filter.Discounted != nil {
		query = query.Where("memberships.discounted = ?", *filter.Discounted)
	}
	if filter.Executive != nil {
		query = query.Where("memberships.executive = ?", *filter.Executive)
	}
	query = r.applyJoinedFilterClauses(query, filter)
	query = filter.Pagination.Apply(query)

	var results []models.MembershipWithAttendance
	if err := query.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *postgresMembershipRepository) Update(membership *models.Membership) error {
	result := r.db.Model(membership).Select("paid", "discounted").Updates(membership)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *postgresMembershipRepository) SetFreeTrialAvailable(id uuid.UUID, available bool) error {
	result := r.db.Model(&models.Membership{}).
		Where("id = ?", id).
		Update("free_trial_available", available)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
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
