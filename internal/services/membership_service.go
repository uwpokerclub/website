package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type membershipService struct {
	db *gorm.DB
}

func NewMembershipService(db *gorm.DB) *membershipService {
	return &membershipService{
		db: db,
	}
}

func (ms *membershipService) CreateMembership(req *models.CreateMembershipRequest) (*models.Membership, error) {
	semesterId, err := uuid.Parse(req.SemesterID)
	if err != nil {
		return nil, e.InvalidRequest("Invalid semester ID specified in request")
	}

	// Create transaction since memberships also affect the semester budget
	tx := ms.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	membership := models.Membership{
		UserID:     req.UserID,
		SemesterID: semesterId,
		Paid:       req.Paid,
		Discounted: req.Discounted,
	}

	res := tx.Create(&membership)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	// Retrieve semester to update the budget
	ss := NewSemesterService(tx)
	semester, err := ss.GetSemester(semesterId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Only update the budget if the membership has been paid for
	if req.Paid {
		// If the membership has been discounted, use the discounted rate instead
		if req.Discounted {
			err = ss.UpdateBudget(semesterId, float64(semester.MembershipDiscountFee))
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			err = ss.UpdateBudget(semesterId, float64(semester.MembershipFee))
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	return &membership, nil
}

func (ms *membershipService) GetMembership(membershipId uuid.UUID) (*models.Membership, error) {
	membership := models.Membership{ID: membershipId}

	res := ms.db.First(&membership)
	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &membership, nil
}

func (ms *membershipService) ListMemberships(filter *models.ListMembershipsFilter) ([]models.ListMembershipsResult, error) {
	ret := []models.ListMembershipsResult{}

	// Find the amount of events each member has participated in the semester
	attendanceQuery := ms.db.
		Select("participants.membership_id, COUNT(*) as total").
		Table("participants").
		Joins("INNER JOIN events ON participants.event_id = events.id").
		Where("events.semester_id = ?", filter.SemesterID).
		Group("participants.membership_id")

	// Get a list of all members in the semester with the number of events they have attended
	// ordered by the members name.
	res := ms.db.
		Select([]string{
			"memberships.id", "users.id as user_id", "users.first_name", "users.last_name",
			"memberships.paid", "memberships.discounted", "COALESCE(attendance.total, 0) as attendance",
		}).
		Table("memberships").
		Joins("LEFT JOIN (?) as attendance ON memberships.id = attendance.membership_id", attendanceQuery).
		Joins("INNER JOIN users ON memberships.user_id = users.id").
		Where("memberships.semester_id = ?", filter.SemesterID).
		Order("users.first_name ASC").
		Order("users.last_name ASC")

	// If a limit is present in the filter, apply this to the query
	if filter.Limit != nil {
		res = res.Limit(*filter.Limit)
	}

	// If an offset is present in the filter, apply this to the query
	if filter.Offset != nil {
		res = res.Offset(*filter.Offset)
	}

	// Fetch the results and return an error if one occured
	res = res.Scan(&ret)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return ret, nil
}

func (ms *membershipService) UpdateMembership(req *models.UpdateMembershipRequest) (*models.Membership, error) {
	// Fetch existing membership
	existingMembership := models.Membership{ID: req.ID}
	res := ms.db.First(&existingMembership)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Validate the request won't put membership into invalid state
	if !req.Paid && req.Discounted {
		return nil, e.InvalidRequest("Cannot set membership to not paid and discounted.")
	}

	// Create transaction
	tx := ms.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Retrieve semester
	ss := NewSemesterService(tx)
	semester, err := ss.GetSemester(existingMembership.SemesterID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Request the membership to be updated to not paid
	if !req.Paid {
		if existingMembership.Paid {
			// Update semester's budget
			if existingMembership.Discounted {
				err = ss.UpdateBudget(existingMembership.SemesterID, -float64(semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			} else {
				err = ss.UpdateBudget(existingMembership.SemesterID, -float64(semester.MembershipFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}
	} else {
		// Requesting to update the member to paid
		// First check if the member is already marked as paid
		if existingMembership.Paid {
			// Next compare the discounted flag
			if !req.Discounted && existingMembership.Discounted {
				// Member is marked as discounted and updating them to not discounted
				err = ss.UpdateBudget(semester.ID, float64(semester.MembershipFee-semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			} else if req.Discounted && !existingMembership.Discounted {
				// Member is not marked as discounted and updating them to discounted
				err = ss.UpdateBudget(semester.ID, -float64(semester.MembershipFee-semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		} else {
			// Existing member has not paid, and we are updating them to paid
			if req.Discounted {
				err = ss.UpdateBudget(semester.ID, float64(semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			} else {
				err = ss.UpdateBudget(semester.ID, float64(semester.MembershipFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}
	}

	existingMembership.Paid = req.Paid
	existingMembership.Discounted = req.Discounted
	tx.Save(&existingMembership)

	if err := tx.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	return &existingMembership, nil
}
