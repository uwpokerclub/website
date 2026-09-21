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
	ErrUpdateLoginNoFields = errors.New("at least one field must be provided")
)

type loginService struct {
	store store.Store
}

// Status-change invariant: any operation that sets a login's status to disabled
// must delete that username's sessions through Sessions().DeleteByUsername in the
// same store transaction. Session roles are snapshots, so leaving those sessions
// active would preserve the account's access until they expire.

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

// UpdateLogin updates a login's password, role, and/or status. At least one field must be provided.
// Returns ErrUpdateLoginNoFields when all inputs are nil and ErrLoginNotFound when the
// username does not exist.
func (svc *loginService) UpdateLogin(username string, password *string, role *string, status *string) error {
	if password == nil && role == nil && status == nil {
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
	if status != nil {
		updates["status"] = *status
	}

	tx, err := svc.store.BeginTx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.Logins().Update(username, updates)
	if errors.Is(err, store.ErrNotFound) {
		return ErrLoginNotFound
	}
	if err != nil {
		return err
	}

	if status != nil && *status == models.LoginStatusDisabled {
		if err := tx.Sessions().DeleteByUsername(username); err != nil {
			return fmt.Errorf("delete sessions: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
