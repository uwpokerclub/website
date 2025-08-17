package testutils

import (
	"api/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateTestSession creates a session for testing authenticated endpoints
func CreateTestSession(db *gorm.DB, username string, role string) (uuid.UUID, error) {
	// Create login record
	login := models.Login{
		Username: username,
		Password: "hashed_password", // In real tests this would be properly hashed
		Role:     role,
	}
	if err := db.Create(&login).Error; err != nil {
		return uuid.Nil, err
	}

	// Create session
	sessionID := uuid.New()
	session := models.Session{
		ID:        sessionID,
		StartedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Username:  username,
	}
	if err := db.Create(&session).Error; err != nil {
		return uuid.Nil, err
	}

	return sessionID, nil
}

// CreateTestUser creates a test user
func CreateTestUser(db *gorm.DB, id uint64, firstName, lastName, email, faculty, questId string) (*models.User, error) {
	user := models.User{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Faculty:   faculty,
		QuestID:   questId,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
