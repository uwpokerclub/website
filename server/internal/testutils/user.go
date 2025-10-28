package testutils

import (
	"api/internal/models"
	"fmt"

	"gorm.io/gorm"
)

var TEST_USERS = []models.User{
	{
		ID:        20780648,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Faculty:   models.FacultyMath,
		QuestID:   "jdoe",
	},
	{
		ID:        20780649,
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Faculty:   models.FacultyEngineering,
		QuestID:   "jsmith",
	},
	{
		ID:        20780650,
		FirstName: "Bob",
		LastName:  "Johnson",
		Email:     "bob.johnson@example.com",
		Faculty:   models.FacultyScience,
		QuestID:   "bjohnson",
	},
}

func SeedUsers(db *gorm.DB) error {
	for _, user := range TEST_USERS {
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}

func FindUserByID(id uint64) (*models.User, error) {
	for _, user := range TEST_USERS {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
