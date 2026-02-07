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
			err = ss.UpdateBudget(semesterId, float32(semester.MembershipDiscountFee))
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			err = ss.UpdateBudget(semesterId, float32(semester.MembershipFee))
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

	res := ms.db.Joins("User").Joins("Semester").First(&membership)
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

func addFilterClauses(query *gorm.DB, filter *models.ListMembershipsFilter) *gorm.DB {
	// Add a WHERE clause for userID if it is present
	if filter.UserID != nil {
		query = query.Where("memberships.user_id = ?", *filter.UserID)
	}

	query = filter.Pagination.Apply(query)

	return query
}

func (ms *membershipService) ListMemberships(filter *models.ListMembershipsFilter) ([]models.ListMembershipsResult, error) {
	ret := []models.ListMembershipsResult{}

	// TODO: Update after participants associations have been setup
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

	res = addFilterClauses(res, filter)

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
				err = ss.UpdateBudget(existingMembership.SemesterID, -float32(semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			} else {
				err = ss.UpdateBudget(existingMembership.SemesterID, -float32(semester.MembershipFee))
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
				err = ss.UpdateBudget(semester.ID, float32(semester.MembershipFee-semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			} else if req.Discounted && !existingMembership.Discounted {
				// Member is not marked as discounted and updating them to discounted
				err = ss.UpdateBudget(semester.ID, -float32(semester.MembershipFee-semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		} else {
			// Existing member has not paid, and we are updating them to paid
			if req.Discounted {
				err = ss.UpdateBudget(semester.ID, float32(semester.MembershipDiscountFee))
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			} else {
				err = ss.UpdateBudget(semester.ID, float32(semester.MembershipFee))
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

func (ms *membershipService) CreateMembershipV2(semesterID uuid.UUID, req *models.CreateMembershipRequestV2) (*models.Membership, error) {
	// Validate the request won't create membership in invalid state
	// Invalid state: paid = false and discounted = true
	if !req.Paid && req.Discounted {
		return nil, errors.New("cannot create membership that is not paid and discounted")
	}

	// Create transaction since memberships also affect the semester budget
	tx := ms.db.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}

	membership := models.Membership{
		UserID:     req.UserID,
		SemesterID: semesterID,
		Paid:       req.Paid,
		Discounted: req.Discounted,
	}

	res := tx.Create(&membership)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Retrieve semester to update the budget
	ss := NewSemesterService(tx)
	semester, err := ss.GetSemester(semesterID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Only update the budget if the membership has been paid for
	if req.Paid {
		// If the membership has been discounted, use the discounted rate instead
		if req.Discounted {
			err = ss.UpdateBudget(semesterID, float32(semester.MembershipDiscountFee))
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			err = ss.UpdateBudget(semesterID, float32(semester.MembershipFee))
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &membership, nil
}

func (ms *membershipService) GetMembershipV2(id uuid.UUID, semesterID uuid.UUID) (*models.Membership, error) {
	var membership models.Membership
	res := models.Membership{}.Preload(ms.db).
		Where("memberships.id = ? AND memberships.semester_id = ?", id, semesterID).
		First(&membership)

	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err := res.Error; err != nil {
		return nil, err
	}

	return &membership, nil
}

func (ms *membershipService) UpdateMembershipV2(id uuid.UUID, semesterID uuid.UUID, req *models.UpdateMembershipRequestV2) (*models.Membership, error) {
	// Fetch existing membership
	var existingMembership models.Membership
	res := models.Membership{}.Preload(ms.db).
		Where("memberships.id = ? AND memberships.semester_id = ?", id, semesterID).
		First(&existingMembership)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, err
	}

	// Validate the request won't put membership into invalid state
	// Determine final state after applying updates
	finalPaid := existingMembership.Paid
	finalDiscounted := existingMembership.Discounted

	if req.Paid != nil {
		finalPaid = *req.Paid
	}
	if req.Discounted != nil {
		finalDiscounted = *req.Discounted
	}

	// Invalid state: paid = false and discounted = true
	if !finalPaid && finalDiscounted {
		return nil, errors.New("cannot set membership to not paid and discounted")
	}

	// Save original values before updating
	originalPaid := existingMembership.Paid
	originalDiscounted := existingMembership.Discounted

	// Create transaction
	tx := ms.db.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Retrieve semester to update the budget
	ss := NewSemesterService(tx)
	semester, err := ss.GetSemester(existingMembership.SemesterID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Determine budget adjustment based on original vs final state
	var budgetAdjustment float32

	// Membership changed from paid to not paid
	if originalPaid && !finalPaid {
		if originalDiscounted {
			budgetAdjustment -= float32(semester.MembershipDiscountFee)
		} else {
			budgetAdjustment -= float32(semester.MembershipFee)
		}
	} else if !originalPaid && finalPaid {
		// Membership changed from not paid to paid
		if finalDiscounted {
			budgetAdjustment += float32(semester.MembershipDiscountFee)
		} else {
			budgetAdjustment += float32(semester.MembershipFee)
		}
	} else if originalPaid && finalPaid {
		// Membership remained paid, check for discount changes
		if originalDiscounted && !finalDiscounted {
			budgetAdjustment += float32(semester.MembershipFee - semester.MembershipDiscountFee)
		} else if !originalDiscounted && finalDiscounted {
			budgetAdjustment -= float32(semester.MembershipFee - semester.MembershipDiscountFee)
		}
	}

	// Apply budget adjustment if needed
	if budgetAdjustment != 0 {
		err = ss.UpdateBudget(existingMembership.SemesterID, budgetAdjustment)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Update the membership with new values
	existingMembership.Paid = finalPaid
	existingMembership.Discounted = finalDiscounted
	tx.Save(&existingMembership)

	if err := tx.Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &existingMembership, nil
}

// ListMembershipsV2 lists all memberships with embedded User and computed attendance count
func (ms *membershipService) ListMembershipsV2(filter *models.ListMembershipsFilter) ([]models.MembershipWithAttendance, int64, error) {
	var memberships []models.Membership

	// Build base query with filters (before pagination)
	baseQuery := ms.db.
		Where("memberships.semester_id = ?", filter.SemesterID)
	baseQuery = addFilterClauses(baseQuery, filter)

	// Get total count before applying pagination
	var total int64
	if err := baseQuery.Model(&models.Membership{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply joins, ordering, and pagination for the actual fetch
	query := baseQuery.Joins("User").
		Order("\"User\".first_name ASC").
		Order("\"User\".last_name ASC")
	query = filter.Pagination.Apply(query)

	if err := query.Find(&memberships).Error; err != nil {
		return nil, 0, err
	}

	// Build attendance map from subquery
	attendanceQuery := ms.db.
		Select("participants.membership_id, COUNT(*) as total").
		Table("participants").
		Joins("INNER JOIN events ON participants.event_id = events.id").
		Where("events.semester_id = ?", filter.SemesterID).
		Group("participants.membership_id")

	type attendanceResult struct {
		MembershipID uuid.UUID
		Total        int
	}
	var attendanceResults []attendanceResult
	if err := attendanceQuery.Scan(&attendanceResults).Error; err != nil {
		return nil, 0, err
	}

	attendanceMap := make(map[uuid.UUID]int)
	for _, ar := range attendanceResults {
		attendanceMap[ar.MembershipID] = ar.Total
	}

	// Build response with attendance
	result := make([]models.MembershipWithAttendance, len(memberships))
	for i, m := range memberships {
		result[i] = models.MembershipWithAttendance{
			Membership: m,
			Attendance: attendanceMap[m.ID],
		}
	}

	return result, total, nil
}
