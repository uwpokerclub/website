package inmemory

import (
	"api/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemberListFiltersByExactQuestID(t *testing.T) {
	repository := NewMemberRepository()
	require.NoError(t, repository.Create(&models.User{ID: 1, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com", Faculty: models.FacultyMath, QuestID: "ada"}))
	require.NoError(t, repository.Create(&models.User{ID: 2, FirstName: "Grace", LastName: "Hopper", Email: "grace@example.com", Faculty: models.FacultyMath, QuestID: "grace"}))

	questID := "ada"
	limit := 2
	members, total, err := repository.List(&models.ListUsersFilter{QuestID: &questID}, &models.Pagination{Limit: &limit})

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, members, 1)
	require.Equal(t, "ada", members[0].QuestID)
}
