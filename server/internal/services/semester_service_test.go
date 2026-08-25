package services

import (
	"api/internal/models"
	"api/internal/store/inmemory"
	"encoding/csv"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSemesterService_UpdateBudget(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semester := &models.Semester{CurrentBudget: 100}
	require.NoError(t, st.Semesters().Create(semester))

	svc := NewSemesterService(st)
	require.NoError(t, svc.UpdateBudget(semester.ID, 25))

	found, err := st.Semesters().FindByID(semester.ID)
	require.NoError(t, err)
	require.InDelta(t, float32(125), found.CurrentBudget, 0.001)
}

func TestSemesterService_UpdateBudget_NotFound(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewSemesterService(st)

	err := svc.UpdateBudget(uuid.New(), 25)
	require.Error(t, err)
}

func TestSemesterService_ExportRankings(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	semesterID := uuid.New()
	membershipID := uuid.New()
	require.NoError(t, st.Rankings().Create(&models.Ranking{
		MembershipID: membershipID,
		Points:       50,
		Membership: &models.Membership{
			ID: membershipID, SemesterID: semesterID,
			User: &models.User{FirstName: "Ada", LastName: "Lovelace"},
		},
	}))

	svc := NewSemesterService(st)
	path, err := svc.ExportRankings(semesterID)
	require.NoError(t, err)
	defer os.Remove(path)

	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	rows, err := csv.NewReader(file).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 2) // header + one ranking
	require.Equal(t, "Ada", rows[1][2])
}
