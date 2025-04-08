package services

import (
	e "api/internal/errors"
	"api/internal/models"

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

func (svc *loginService) CreateLogin(username string, password string, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	login := models.Login{
		Username: username,
		Password: string(hash),
		Role:     role,
	}

	res := svc.db.Create(&login)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}
