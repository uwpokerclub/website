package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
)

type semesterService struct {
	store store.Store
}

func NewSemesterService(st store.Store) *semesterService {
	return &semesterService{
		store: st,
	}
}

func (ss *semesterService) UpdateBudget(id uuid.UUID, amount float32) error {
	err := ss.store.Semesters().IncrementBudget(id, amount)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return e.NotFound("semester not found")
		}
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (ss *semesterService) ExportRankings(id uuid.UUID) (string, error) {
	// Get the top 100 rankings for export (limited to prevent excessive file sizes)
	limit := 100
	rankings, _, err := ss.store.Rankings().List(&models.ListRankingsFilter{
		SemesterID: id,
		Pagination: models.Pagination{Limit: &limit},
	})
	if err != nil {
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
