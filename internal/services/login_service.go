package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type loginService struct {
	db *gorm.DB
}

func NewLoginService(db *gorm.DB) *loginService {
	return &loginService{
		db: db,
	}
}

func (svc *loginService) CreateLogin(username string, password string) error {
	existingLogins := []models.Login{}
	res := svc.db.Model(&models.Login{}).Find(&existingLogins)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	if len(existingLogins) > 0 {
		return e.Forbidden("You cannot perform this action")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	login := models.Login{
		Username: username,
		Password: string(hash),
	}

	res = svc.db.Create(&login)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

// TODO: Remove once session based authentication is implemented on app and api side
func (svc *loginService) ValidateCredentials(username string, password string) (string, error) {
	login := models.Login{Username: username}

	res := svc.db.First(&login)
	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return "", e.Unauthorized("Invalid username or password")
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return "", e.InternalServerError(err.Error())
	}

	// Compare hashed password with plain text version
	err := bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(password))
	if err != nil {
		return "", e.Unauthorized("Invalid username or password")
	}

	// Generate a new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	secretKey := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", e.InternalServerError(fmt.Sprintf("Error occurred signing token: %v", err))
	}

	return tokenString, nil
}
