package services

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"github.com/google/uuid"
)

type membershipService struct {
	store store.Store
}

func NewMembershipService(st store.Store) *membershipService {
	return &membershipService{
		store: st,
	}
}

func (ms *membershipService) CreateMembership(semesterID uuid.UUID, req *models.CreateMembershipRequest) (*models.Membership, error) {
	if !req.Paid && req.Discounted {
		return nil, errors.New("cannot create membership that is not paid and discounted")
	}

	tx, err := ms.store.BeginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	membership := models.Membership{
		UserID:     req.UserID,
		SemesterID: semesterID,
		Paid:       req.Paid,
		Discounted: req.Discounted,
	}

	if err := tx.Memberships().Create(&membership); err != nil {
		return nil, err
	}

	semester, err := tx.Semesters().FindByID(semesterID)
	if err != nil {
		return nil, err
	}

	if req.Paid {
		fee := semester.MembershipFee
		if req.Discounted {
			fee = semester.MembershipDiscountFee
		}
		if err := tx.Semesters().IncrementBudget(semesterID, float32(fee)); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &membership, nil
}

func (ms *membershipService) UpdateMembership(id uuid.UUID, semesterID uuid.UUID, req *models.UpdateMembershipRequest) (*models.Membership, error) {
	existingMembership, err := ms.store.Memberships().FindByIDAndSemesterID(id, semesterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	finalPaid := existingMembership.Paid
	finalDiscounted := existingMembership.Discounted
	if req.Paid != nil {
		finalPaid = *req.Paid
	}
	if req.Discounted != nil {
		finalDiscounted = *req.Discounted
	}

	if !finalPaid && finalDiscounted {
		return nil, errors.New("cannot set membership to not paid and discounted")
	}

	originalPaid := existingMembership.Paid
	originalDiscounted := existingMembership.Discounted

	tx, err := ms.store.BeginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	semester, err := tx.Semesters().FindByID(existingMembership.SemesterID)
	if err != nil {
		return nil, err
	}

	var budgetAdjustment float32
	if originalPaid && !finalPaid {
		if originalDiscounted {
			budgetAdjustment -= float32(semester.MembershipDiscountFee)
		} else {
			budgetAdjustment -= float32(semester.MembershipFee)
		}
	} else if !originalPaid && finalPaid {
		if finalDiscounted {
			budgetAdjustment += float32(semester.MembershipDiscountFee)
		} else {
			budgetAdjustment += float32(semester.MembershipFee)
		}
	} else if originalPaid && finalPaid {
		if originalDiscounted && !finalDiscounted {
			budgetAdjustment += float32(semester.MembershipFee - semester.MembershipDiscountFee)
		} else if !originalDiscounted && finalDiscounted {
			budgetAdjustment -= float32(semester.MembershipFee - semester.MembershipDiscountFee)
		}
	}

	if budgetAdjustment != 0 {
		if err := tx.Semesters().IncrementBudget(existingMembership.SemesterID, budgetAdjustment); err != nil {
			return nil, err
		}
	}

	existingMembership.Paid = finalPaid
	existingMembership.Discounted = finalDiscounted
	if err := tx.Memberships().Update(&existingMembership); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &existingMembership, nil
}
