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
