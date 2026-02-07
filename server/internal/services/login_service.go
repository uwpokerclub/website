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

// loginWithUserRow is an intermediate struct for scanning LEFT JOIN results
type loginWithUserRow struct {
	Username  string
	Role      string
	UserID    *uint64
	FirstName *string
	LastName  *string
}

// ListLogins retrieves all logins with their linked member information
func (svc *loginService) ListLogins(pagination *models.Pagination) ([]models.LoginWithMember, int64, error) {
	// Count total before pagination
	var total int64
	if err := svc.db.Table("logins").Count(&total).Error; err != nil {
		return nil, 0, e.InternalServerError(err.Error())
	}

	var rows []loginWithUserRow

	// Use LEFT JOIN to fetch logins with linked members, excluding password
	query := svc.db.Table("logins").
		Select("logins.username, logins.role, users.id as user_id, users.first_name, users.last_name").
		Joins("LEFT JOIN users ON logins.username = users.quest_id").
		Order("logins.username ASC")
	query = pagination.Apply(query)

	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, e.InternalServerError(err.Error())
	}

	// Transform to LoginWithMember
	results := make([]models.LoginWithMember, len(rows))
	for i, row := range rows {
		results[i] = models.LoginWithMember{
			Username: row.Username,
			Role:     row.Role,
		}
		if row.UserID != nil && *row.UserID != 0 {
			results[i].LinkedMember = &models.LinkedMemberInfo{
				ID:        *row.UserID,
				FirstName: *row.FirstName,
				LastName:  *row.LastName,
			}
		}
	}

	return results, total, nil
}

// GetLogin retrieves a single login by username (without password)
func (svc *loginService) GetLogin(username string) (*models.LoginWithMember, error) {
	var row loginWithUserRow

	// Use LEFT JOIN to fetch login with linked member, excluding password
	err := svc.db.Table("logins").
		Select("logins.username, logins.role, users.id as user_id, users.first_name, users.last_name").
		Joins("LEFT JOIN users ON logins.username = users.quest_id").
		Where("logins.username = ?", username).
		Scan(&row).Error

	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Check if login was found
	if row.Username == "" {
		return nil, e.NotFound("login not found")
	}

	// Transform to LoginWithMember
	result := &models.LoginWithMember{
		Username: row.Username,
		Role:     row.Role,
	}
	if row.UserID != nil && *row.UserID != 0 {
		result.LinkedMember = &models.LinkedMemberInfo{
			ID:        *row.UserID,
			FirstName: *row.FirstName,
			LastName:  *row.LastName,
		}
	}

	return result, nil
}

// DeleteLogin deletes a login by username
func (svc *loginService) DeleteLogin(username string) error {
	// Delete the login and check rows affected (sessions will cascade delete)
	res := svc.db.Where("username = ?", username).Delete(&models.Login{})
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	if res.RowsAffected == 0 {
		return e.NotFound("login not found")
	}

	return nil
}

// ChangePassword changes the password for a login
func (svc *loginService) ChangePassword(username string, newPassword string) error {
	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return e.InternalServerError(err.Error())
	}

	// Update the password and check rows affected
	res := svc.db.Model(&models.Login{}).Where("username = ?", username).Update("password", string(hash))
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	if res.RowsAffected == 0 {
		return e.NotFound("login not found")
	}

	return nil
}

// CreateLoginFromRequest creates a new login from a CreateLoginRequest
func (svc *loginService) CreateLoginFromRequest(req *models.CreateLoginRequest) error {
	return svc.CreateLogin(req.Username, req.Password, req.Role)
}
