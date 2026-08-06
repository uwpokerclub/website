package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresRankingRepository struct {
	db *gorm.DB
}

var _ store.RankingRepository = (*postgresRankingRepository)(nil)

func NewRankingRepository(db *gorm.DB) store.RankingRepository {
	return &postgresRankingRepository{db: db}
}

func (r *postgresRankingRepository) Create(ranking *models.Ranking) error {
	return r.db.Create(ranking).Error
}

func (r *postgresRankingRepository) FindByMembershipID(membershipID uuid.UUID) (models.Ranking, error) {
	var ranking models.Ranking

	err := r.db.Where("membership_id = ?", membershipID).First(&ranking).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Ranking{}, store.ErrNotFound
		}
		return models.Ranking{}, err
	}

	return ranking, nil
}

func (r *postgresRankingRepository) Update(ranking *models.Ranking) error {
	return r.db.Model(ranking).Select("points").Updates(ranking).Error
}

func (r *postgresRankingRepository) BatchIncrementPoints(updates map[uuid.UUID]int32) error {
	if len(updates) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)*2)
	i := 1
	for membershipID, points := range updates {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d::uuid, $%d, 0)", i, i+1))
		args = append(args, membershipID, points)
		i += 2
	}

	query := fmt.Sprintf(
		`INSERT INTO rankings (membership_id, points, attendance) VALUES %s ON CONFLICT (membership_id) DO UPDATE SET points = rankings.points + EXCLUDED.points`,
		strings.Join(valueStrings, ", "),
	)

	return r.db.Exec(query, args...).Error
}

func (r *postgresRankingRepository) FindBySemesterAndMembershipID(semesterID uuid.UUID, membershipID uuid.UUID) (models.GetRankingResponse, error) {
	var ret models.GetRankingResponse

	err := r.db.
		Table(models.SemesterRankingsView).
		Select("points", "position").
		Where("semester_id = ? AND membership_id = ?", semesterID, membershipID).
		First(&ret).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.GetRankingResponse{}, store.ErrNotFound
		}
		return models.GetRankingResponse{}, err
	}

	return ret, nil
}

func (r *postgresRankingRepository) List(filter *models.ListRankingsFilter) ([]models.RankingResponse, int64, error) {
	base := func() *gorm.DB {
		q := r.db.Table(models.SemesterRankingsView).Where("semester_id = ?", filter.SemesterID)
		if filter.Search != "" {
			pattern := "%" + eventNameLikeReplacer.Replace(filter.Search) + "%"
			q = q.Where(
				"first_name ILIKE ? OR last_name ILIKE ? OR (first_name || ' ' || last_name) ILIKE ?",
				pattern, pattern, pattern,
			)
		}
		return q
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rankings []models.RankingResponse
	query := base().
		Select("user_id as id, first_name, last_name, points, position").
		Order("position ASC, last_name ASC, first_name ASC")
	query = filter.Pagination.Apply(query)

	if err := query.Find(&rankings).Error; err != nil {
		return nil, 0, err
	}

	return rankings, total, nil
}
