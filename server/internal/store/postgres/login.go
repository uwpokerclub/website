package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"

	"gorm.io/gorm"
)

type postgresLoginRepository struct {
	db *gorm.DB
}

var _ store.LoginRepository = (*postgresLoginRepository)(nil)

func NewLoginRepository(db *gorm.DB) store.LoginRepository {
	return &postgresLoginRepository{db: db}
}

func (r *postgresLoginRepository) Create(login *models.Login) error {
	return r.db.Create(login).Error
}

func (r *postgresLoginRepository) FindByUsername(username string) (models.Login, error) {
	var login models.Login

	err := r.db.Where("username = ?", username).First(&login).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Login{}, store.ErrNotFound
		}
		return models.Login{}, err
	}

	return login, nil
}

func (r *postgresLoginRepository) Update(username string, values map[string]any) error {
	result := r.db.Model(&models.Login{}).Where("username = ?", username).Updates(values)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (r *postgresLoginRepository) Delete(username string) error {
	result := r.db.Where("username = ?", username).Delete(&models.Login{})
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

type loginWithUserRow struct {
	Username  string
	Role      string
	UserID    *uint64
	FirstName *string
	LastName  *string
}

func toLoginWithMember(row loginWithUserRow) models.LoginWithMember {
	result := models.LoginWithMember{Username: row.Username, Role: row.Role}
	if row.UserID != nil && *row.UserID != 0 {
		result.LinkedMember = &models.LinkedMemberInfo{
			ID:        *row.UserID,
			FirstName: *row.FirstName,
			LastName:  *row.LastName,
		}
	}
	return result
}

func (r *postgresLoginRepository) List(pagination *models.Pagination, search string) ([]models.LoginWithMember, int64, error) {
	base := func() *gorm.DB {
		q := r.db.Table("logins").Joins("LEFT JOIN users ON logins.username = users.quest_id")
		if search != "" {
			pattern := "%" + eventNameLikeReplacer.Replace(search) + "%"
			q = q.Where(
				"logins.username ILIKE ? OR logins.role ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR (users.first_name || ' ' || users.last_name) ILIKE ?",
				pattern, pattern, pattern, pattern, pattern,
			)
		}
		return q
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []loginWithUserRow
	query := base().
		Select("logins.username, logins.role, users.id as user_id, users.first_name, users.last_name").
		Order("logins.username ASC")
	query = pagination.Apply(query)

	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	results := make([]models.LoginWithMember, len(rows))
	for i, row := range rows {
		results[i] = toLoginWithMember(row)
	}

	return results, total, nil
}

func (r *postgresLoginRepository) FindByUsernameWithMember(username string) (models.LoginWithMember, error) {
	var row loginWithUserRow

	err := r.db.Table("logins").
		Select("logins.username, logins.role, users.id as user_id, users.first_name, users.last_name").
		Joins("LEFT JOIN users ON logins.username = users.quest_id").
		Where("logins.username = ?", username).
		Scan(&row).Error
	if err != nil {
		return models.LoginWithMember{}, err
	}

	if row.Username == "" {
		return models.LoginWithMember{}, store.ErrNotFound
	}

	return toLoginWithMember(row), nil
}
