package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rankingService struct {
	db *gorm.DB
}

func NewRankingService(db *gorm.DB) *rankingService {
	return &rankingService{
		db: db,
	}
}

func (svc *rankingService) UpdateRanking(membershipId uuid.UUID, points int) error {
	var ranking models.Ranking

	res := svc.db.Where("membership_id = ?", membershipId).First(&ranking)
	// If the record can not be found, create a new ranking record
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		ranking = models.Ranking{
			MembershipID: membershipId,
			Points:       int32(points),
		}

		res = svc.db.Create(&ranking)
		if err := res.Error; err != nil {
			return e.InternalServerError(err.Error())
		}

		return nil
	}
	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	ranking.Points += int32(points)

	res = svc.db.Save(&ranking)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *rankingService) BatchUpdateRankings(updates map[uuid.UUID]int) error {
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

	if err := svc.db.Exec(query, args...).Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *rankingService) GetRanking(semesterID uuid.UUID, membershipID uuid.UUID) (*models.GetRankingResponse, error) {
	ret := models.GetRankingResponse{}

	res := svc.db.
		Table(models.SemesterRankingsView).
		Select("points", "position").
		Where("semester_id = ? AND membership_id = ?", semesterID, membershipID).
		First(&ret)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &ret, nil
}
