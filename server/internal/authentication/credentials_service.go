package authentication

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type credentialsService struct {
	db *gorm.DB
}

func NewCredentialService(db *gorm.DB) *credentialsService {
	return &credentialsService{
		db: db,
	}
}

func (svc *credentialsService) Validate(username string, password string) (bool, error) {
	login := models.Login{Username: username}

	// Find first login with the specified username
	res := svc.db.First(&login)
	// Check if the login was found
	err := res.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	// If the error is not a not found error,
	// then return this error as a server error
	if err != nil {
		return false, e.InternalServerError(err.Error())
	}

	// Compare the hashed password and the plaintext password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}
