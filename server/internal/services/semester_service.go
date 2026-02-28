package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
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

func (ss *semesterService) ListSemestersV2(pagination *models.Pagination) ([]models.Semester, int64, error) {
	var total int64
	if err := ss.db.Model(&models.Semester{}).Count(&total).Error; err != nil {
		return nil, 0, e.InternalServerError(err.Error())
	}

	var semesters []models.Semester
	query := ss.db.Order("start_date DESC")
	query = pagination.Apply(query)

	if err := query.Find(&semesters).Error; err != nil {
		return nil, 0, e.InternalServerError(err.Error())
	}

	return semesters, total, nil
}

func (ss *semesterService) GetRankings(id uuid.UUID) ([]models.RankingResponse, error) {
	var rankings []models.RankingResponse

	res := ss.db.
		Table(models.SemesterRankingsView).
		Select("user_id as id, first_name, last_name, points, position").
		Where("semester_id = ?", id).
		Order("position ASC, last_name ASC, first_name ASC").
		Find(&rankings)

	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return rankings, nil
}

func (ss *semesterService) GetRankingsV2(id uuid.UUID, pagination *models.Pagination) ([]models.RankingResponse, int64, error) {
	var total int64
	if err := ss.db.Table(models.SemesterRankingsView).Where("semester_id = ?", id).Count(&total).Error; err != nil {
		return nil, 0, e.InternalServerError(err.Error())
	}

	var rankings []models.RankingResponse
	query := ss.db.
		Table(models.SemesterRankingsView).
		Select("user_id as id, first_name, last_name, points, position").
		Where("semester_id = ?", id).
		Order("position ASC, last_name ASC, first_name ASC")
	query = pagination.Apply(query)

	if err := query.Find(&rankings).Error; err != nil {
		return nil, 0, e.InternalServerError(err.Error())
	}

	return rankings, total, nil
}

func (ss *semesterService) UpdateBudget(id uuid.UUID, amount float32) error {
	// Use atomic update to prevent race conditions
	res := ss.db.Model(&models.Semester{}).
		Where("id = ?", id).
		Update("current_budget", gorm.Expr("current_budget + ?", amount))

	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	// Check if the semester was found (RowsAffected will be 0 if not found)
	if res.RowsAffected == 0 {
		return e.NotFound("semester not found")
	}

	return nil
}

func (ss *semesterService) ExportRankings(id uuid.UUID) (string, error) {
	var rankings []models.RankingResponse

	// Get the top 100 rankings for export (limited to prevent excessive file sizes)
	res := ss.db.
		Table(models.SemesterRankingsView).
		Select("user_id as id, first_name, last_name, points, position").
		Where("semester_id = ?", id).
		Order("position ASC, last_name ASC, first_name ASC").
		Limit(100).
		Find(&rankings)

	if err := res.Error; err != nil {
		return "", e.InternalServerError(fmt.Sprintf("Error when retrieving rankings: %s", err.Error()))
	}

	// Open a new CSV file in the OS temp directory
	file, err := os.CreateTemp("", "rankings-*.csv")
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
	writer.Write([]string{"position", "id", "first_name", "last_name", "points"})

	// Write each row from the database to the CSV
	for _, ranking := range rankings {
		record := []string{
			strconv.FormatInt(int64(ranking.Position), 10),
			strconv.FormatUint(ranking.ID, 10),
			ranking.FirstName,
			ranking.LastName,
			strconv.FormatInt(int64(ranking.Points), 10),
		}

		if err := writer.Write(record); err != nil {
			return "", e.InternalServerError(fmt.Sprintf("Error writing to the CSV file: %s", err.Error()))
		}
	}

	return file.Name(), nil
}
