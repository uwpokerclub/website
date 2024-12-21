package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type semesterService struct {
	db *gorm.DB
}

func NewSemesterService(db *gorm.DB) *semesterService {
	return &semesterService{
		db: db,
	}
}

func (ss *semesterService) CreateSemester(req *models.CreateSemesterRequest) (*models.Semester, error) {
	semester := models.Semester{
		Name:                  req.Name,
		Meta:                  req.Meta,
		StartDate:             req.StartDate,
		EndDate:               req.EndDate,
		StartingBudget:        req.StartingBudget,
		CurrentBudget:         req.StartingBudget,
		MembershipFee:         req.MembershipFee,
		MembershipDiscountFee: req.MembershipDiscountFee,
		RebuyFee:              req.RebuyFee,
	}

	res := ss.db.Create(&semester)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &semester, nil
}

func (ss *semesterService) GetSemester(id uuid.UUID) (*models.Semester, error) {
	semester := models.Semester{ID: id}

	res := ss.db.First(&semester)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &semester, nil
}

func (ss *semesterService) ListSemesters() ([]models.Semester, error) {
	var semesters []models.Semester

	res := ss.db.Order("start_date DESC").Find(&semesters)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return semesters, nil
}

func (ss *semesterService) GetRankings(id uuid.UUID) ([]models.RankingResponse, error) {
	var rankings []models.RankingResponse

	res := ss.db.
		Table("memberships").
		Select("users.id, users.first_name, users.last_name, rankings.points").
		Joins("INNER JOIN users ON memberships.user_id = users.id").
		Joins("INNER JOIN rankings ON memberships.id = rankings.membership_id").
		Where("memberships.semester_id = ?", id).
		Order("rankings.points DESC").
		Find(&rankings)

	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return rankings, nil
}

func (ss *semesterService) UpdateBudget(id uuid.UUID, amount float64) error {
	semester := models.Semester{ID: id}

	res := ss.db.First(&semester)

	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	// Update budget by the new amount
	semester.CurrentBudget += amount

	res = ss.db.Save(&semester)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (ss *semesterService) ExportRankings(id uuid.UUID) (string, error) {
	var rankings []models.RankingResponse

	// Get the top 100 rankings
	res := ss.db.
		Table("memberships").
		Select("users.id, users.first_name, users.last_name, rankings.points").
		Joins("INNER JOIN users ON memberships.user_id = users.id").
		Joins("INNER JOIN rankings on memberships.id = rankings.membership_id").
		Where("memberships.semester_id = ?", id).
		Order("rankings.points DESC").
		Limit(100).
		Find(&rankings)

	if err := res.Error; err != nil {
		return "", e.InternalServerError(fmt.Sprintf("Error when retrieving rankings: %s", err.Error()))
	}

	// Open a new CSV file
	filename := "rankings.csv"
	file, err := os.Create(filename)
	if err != nil {
		return "", e.InternalServerError(fmt.Sprintf("Error when creating rankings file: %s", err.Error()))
	}
	// Ensure the file is closed at the end of the function
	defer file.Close()

	// Initialize a new CSV writer
	writer := csv.NewWriter(file)
	// Ensure all data is written to the file
	defer writer.Flush()

	// Write headers to the file
	writer.Write([]string{"id", "first_name", "last_name", "points"})

	// Write each row from the database to the CSV
	for _, ranking := range rankings {
		record := []string{
			strconv.FormatUint(ranking.ID, 10),
			ranking.FirstName,
			ranking.LastName,
			strconv.FormatInt(int64(ranking.Points), 10),
		}

		if err := writer.Write(record); err != nil {
			return "", e.InternalServerError(fmt.Sprintf("Error writing to the CSV file: %s", err.Error()))
		}
	}

	// Get the absolute filepath to the new file
	fp, err := filepath.Abs(filename)
	if err != nil {
		return "", e.InternalServerError(fmt.Sprintf("Failed to export CSV"))
	}

	return fp, nil
}
