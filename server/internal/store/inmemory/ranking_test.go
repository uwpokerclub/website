package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRankingRepository_Create(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()

	ranking := &models.Ranking{MembershipID: uuid.New(), Points: 10}
	require.NoError(t, repo.Create(ranking))
	require.NotZero(t, ranking.ID)
}

func TestRankingRepository_Create_DuplicateMembership(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	membershipID := uuid.New()

	require.NoError(t, repo.Create(&models.Ranking{MembershipID: membershipID, Points: 5}))

	err := repo.Create(&models.Ranking{MembershipID: membershipID, Points: 1})
	require.Error(t, err)
}

func TestRankingRepository_FindByMembershipID(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	membershipID := uuid.New()

	created := &models.Ranking{MembershipID: membershipID, Points: 15}
	require.NoError(t, repo.Create(created))

	found, err := repo.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.EqualValues(t, 15, found.Points)

	_, err = repo.FindByMembershipID(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestRankingRepository_Update(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	membershipID := uuid.New()

	created := &models.Ranking{MembershipID: membershipID, Points: 5}
	require.NoError(t, repo.Create(created))

	created.Points = 20
	require.NoError(t, repo.Update(created))

	found, err := repo.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.EqualValues(t, 20, found.Points)
}

func TestRankingRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()

	err := repo.Update(&models.Ranking{ID: 99999, Points: 1})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestRankingRepository_BatchIncrementPoints_CreatesForNewMembership(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	membershipID := uuid.New()

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{membershipID: 12}))

	found, err := repo.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.EqualValues(t, 12, found.Points)
}

func TestRankingRepository_BatchIncrementPoints_IncrementsExisting(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	membershipID := uuid.New()
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: membershipID, Points: 8}))

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{membershipID: 5}))

	found, err := repo.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.EqualValues(t, 13, found.Points)
}

func TestRankingRepository_BatchIncrementPoints_Mixed(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	existingID := uuid.New()
	newID := uuid.New()
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: existingID, Points: 3}))

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{
		existingID: 7,
		newID:      4,
	}))

	foundExisting, err := repo.FindByMembershipID(existingID)
	require.NoError(t, err)
	require.EqualValues(t, 10, foundExisting.Points)

	foundNew, err := repo.FindByMembershipID(newID)
	require.NoError(t, err)
	require.EqualValues(t, 4, foundNew.Points)
}

func TestRankingRepository_BatchIncrementPoints_Empty(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()

	require.NoError(t, repo.BatchIncrementPoints(map[uuid.UUID]int32{}))
}

func TestRankingRepository_Clone_Isolation(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()
	membershipID := uuid.New()

	ranking := &models.Ranking{MembershipID: membershipID, Points: 5}
	require.NoError(t, repo.Create(ranking))

	clone := repo.clone()

	// Mutating the clone must not affect the original.
	require.NoError(t, clone.Update(&models.Ranking{ID: ranking.ID, MembershipID: membershipID, Points: 99}))

	original, err := repo.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.EqualValues(t, 5, original.Points)

	cloned, err := clone.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.EqualValues(t, 99, cloned.Points)

	// Creating a new ranking on the original must not appear in the clone.
	require.NoError(t, repo.Create(&models.Ranking{MembershipID: uuid.New(), Points: 1}))

	_, err = clone.FindByMembershipID(membershipID)
	require.NoError(t, err)
	require.Len(t, clone.rankings, 1)
	require.Len(t, repo.rankings, 2)
}

func TestRankingRepository_FindBySemesterAndMembershipID(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()

	semesterID := uuid.New()
	membershipID := uuid.New()
	require.NoError(t, repo.Create(&models.Ranking{
		MembershipID: membershipID,
		Points:       42,
		Membership:   &models.Membership{ID: membershipID, SemesterID: semesterID},
	}))

	found, err := repo.FindBySemesterAndMembershipID(semesterID, membershipID)
	require.NoError(t, err)
	require.EqualValues(t, 42, found.Points)
	require.EqualValues(t, 1, found.Position)

	_, err = repo.FindBySemesterAndMembershipID(semesterID, uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestRankingRepository_List(t *testing.T) {
	t.Parallel()

	repo := newRankingRepository()

	semesterID := uuid.New()
	m1, m2 := uuid.New(), uuid.New()
	require.NoError(t, repo.Create(&models.Ranking{
		MembershipID: m1, Points: 10,
		Membership: &models.Membership{ID: m1, SemesterID: semesterID, User: &models.User{FirstName: "Rosalind", LastName: "Franklin"}},
	}))
	require.NoError(t, repo.Create(&models.Ranking{
		MembershipID: m2, Points: 20,
		Membership: &models.Membership{ID: m2, SemesterID: semesterID, User: &models.User{FirstName: "Marie", LastName: "Curie"}},
	}))

	results, total, err := repo.List(&models.ListRankingsFilter{SemesterID: semesterID})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, results, 2)
	// Higher points ranks first (position 1).
	require.Equal(t, "Marie", results[0].FirstName)
	require.EqualValues(t, 1, results[0].Position)
	require.Equal(t, "Rosalind", results[1].FirstName)
	require.EqualValues(t, 2, results[1].Position)

	results, total, err = repo.List(&models.ListRankingsFilter{SemesterID: semesterID, Search: "franklin"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "Rosalind", results[0].FirstName)
}
