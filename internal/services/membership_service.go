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

func (ms *membershipService) ListMemberships(semesterId uuid.UUID) ([]models.ListMembershipsResult, error) {
	var ret []models.ListMembershipsResult

	res := ms.db.Raw(`
	WITH attendance AS (
		SELECT
			p.membership_id,
			COUNT(*) AS total
		FROM
			participants p
			INNER JOIN events e ON p.event_id = e.id
		WHERE
			e.semester_id = $1
		GROUP BY
			p.membership_id
		)
		
		SELECT
			m.id,
			users.id AS user_id,
			users.first_name,
			users.last_name,
			m.paid,
			m.discounted,
			coalesce(a.total, 0) AS attendance
		FROM
			memberships m
			INNER JOIN users ON m.user_id = users.id
			left JOIN attendance a ON m.id = a.membership_id
		WHERE m.semester_id = $1
		ORDER BY
			users.first_name ASC,
			users.last_name ASC;
	`, semesterId).Find(&ret)

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
