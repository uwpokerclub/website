package authentication

import (
	e "api/internal/errors"
	"api/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionManager struct {
	db *gorm.DB
}

func NewSessionManager(db *gorm.DB) *sessionManager {
	return &sessionManager{
		db: db,
	}
}

func (svc *sessionManager) Create(username string) (uuid.UUID, error) {
	// Get the current time
	now := time.Now().UTC()
	// Set the expiry time to 8 hours in the future
	expiry := now.Add(time.Hour * 8).UTC()

	// Create the session in the database
	session := models.Session{StartedAt: time.Now(), ExpiresAt: expiry, Username: username}
	res := svc.db.Create(&session)

	if err := res.Error; err != nil {
		return uuid.UUID{}, e.InternalServerError(err.Error())
	}

	return session.ID, nil
}

func (svc *sessionManager) Invalidate(sessionID uuid.UUID) error {
	session := models.Session{ID: sessionID}

	res := svc.db.Delete(&session)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}
