package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"

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
	ranking := models.Ranking{
		MembershipID: membershipId,
	}

	res := svc.db.First(&ranking)
	// If the record can not be found, create a new ranking record
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		ranking.Points = int32(points)

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

	res = svc.db.Where("membership_id = ?", membershipId).Save(&ranking)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *rankingService) GetRanking(semesterID uuid.UUID, membershipID uuid.UUID) (*models.GetRankingResponse, error) {
	ret := models.GetRankingResponse{}

	rankingsTableQuery := svc.db.
		Table("memberships").
		Select("rankings.membership_id, rankings.points, ROW_NUMBER() OVER (ORDER BY rankings.points DESC) as position").
		Joins("INNER JOIN rankings ON memberships.id = rankings.membership_id").
		Where("memberships.semester_id = ?", semesterID)

	res := svc.db.
		Table("(?) as rankings", rankingsTableQuery).
		Select("points", "position").
		Where("membership_id = ?", membershipID).
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
