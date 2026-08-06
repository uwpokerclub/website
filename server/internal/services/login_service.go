package services

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors returned by loginService. Controllers map these to HTTP responses.
var (
	ErrLoginNotFound       = errors.New("login not found")
	ErrUpdateLoginNoFields = errors.New("at least one of password or role must be provided")
)

type loginService struct {
	store store.Store
}

func NewLoginService(st store.Store) *loginService {
	return &loginService{
		store: st,
	}
}

func (svc *loginService) CreateLogin(username string, password string, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	login := models.Login{
		Username: username,
		Password: string(hash),
		Role:     role,
	}

	return svc.store.Logins().Create(&login)
}

// UpdateLogin updates a login's password and/or role. At least one field must be provided.
// Returns ErrUpdateLoginNoFields when both inputs are nil and ErrLoginNotFound when the
// username does not exist.
func (svc *loginService) UpdateLogin(username string, password *string, role *string) error {
	if password == nil && role == nil {
		return ErrUpdateLoginNoFields
	}

	updates := map[string]any{}

	if password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		updates["password"] = string(hash)
	}

	if role != nil {
		updates["role"] = *role
	}

	err := svc.store.Logins().Update(username, updates)
	if errors.Is(err, store.ErrNotFound) {
		return ErrLoginNotFound
	}

	return err
}
