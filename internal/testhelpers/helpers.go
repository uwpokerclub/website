package testhelpers

import (
	"api/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateSemester(db *gorm.DB, id uuid.UUID, name string, meta string, startDate time.Time, endDate time.Time, startingBudget float64, currentBudget float64, membershipFee int8, discountFee int8, rebuyFee int8) (*models.Semester, error) {
	semester := models.Semester{
		ID:                    id,
		Name:                  name,
		Meta:                  meta,
		StartDate:             startDate,
		EndDate:               endDate,
		StartingBudget:        startingBudget,
		CurrentBudget:         currentBudget,
		MembershipFee:         membershipFee,
		MembershipFeeDiscount: discountFee,
		RebuyFee:              rebuyFee,
	}

	res := db.Create(&semester)
	if res.Error != nil {
		return nil, res.Error
	}

	return &semester, nil
}

func CreateUser(db *gorm.DB, id uint64, firstName string, lastName string, email string, faculty string, questId string) (*models.User, error) {
	user := models.User{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Faculty:   faculty,
		QuestID:   questId,
	}

	res := db.Create(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}
