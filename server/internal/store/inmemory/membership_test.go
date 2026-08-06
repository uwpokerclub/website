package inmemory

import (
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMembershipRepository_CreateAndFindByID(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	membership := &models.Membership{
		UserID:     1,
		SemesterID: uuid.New(),
		Paid:       true,
		Discounted: false,
	}

	require.NoError(t, repo.Create(membership))
	require.NotEqual(t, uuid.Nil, membership.ID)

	found, err := repo.FindByID(membership.ID)
	require.NoError(t, err)
	require.Equal(t, *membership, found)
}

func TestMembershipRepository_Create_DuplicateID(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	id := uuid.New()
	require.NoError(t, repo.Create(&models.Membership{ID: id, UserID: 1, SemesterID: uuid.New()}))

	err := repo.Create(&models.Membership{ID: id, UserID: 2, SemesterID: uuid.New()})
	require.Error(t, err)
}

func TestMembershipRepository_Create_DuplicateUserSemester(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	semesterID := uuid.New()
	require.NoError(t, repo.Create(&models.Membership{UserID: 1, SemesterID: semesterID}))

	err := repo.Create(&models.Membership{UserID: 1, SemesterID: semesterID})
	require.Error(t, err)
}

func TestMembershipRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	_, err := repo.FindByID(uuid.New())
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_FindByIDAndSemesterID(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	semesterA := uuid.New()
	semesterB := uuid.New()

	membership := &models.Membership{UserID: 1, SemesterID: semesterA}
	require.NoError(t, repo.Create(membership))

	found, err := repo.FindByIDAndSemesterID(membership.ID, semesterA)
	require.NoError(t, err)
	require.Equal(t, membership.ID, found.ID)

	_, err = repo.FindByIDAndSemesterID(membership.ID, semesterB)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_List(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	semesterA := uuid.New()
	semesterB := uuid.New()

	m1 := &models.Membership{UserID: 1, SemesterID: semesterA, Paid: true, User: &models.User{ID: 1, FirstName: "Alice", LastName: "A", Email: "alice@example.com"}}
	m2 := &models.Membership{UserID: 2, SemesterID: semesterA, Paid: false, User: &models.User{ID: 2, FirstName: "Bob", LastName: "B", Email: "bob@example.com"}}
	m3 := &models.Membership{UserID: 3, SemesterID: semesterB, Paid: true, User: &models.User{ID: 3, FirstName: "Cara", LastName: "C", Email: "cara@example.com"}}
	require.NoError(t, repo.Create(m1))
	require.NoError(t, repo.Create(m2))
	require.NoError(t, repo.Create(m3))

	results, total, err := repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, results, 2)
	require.Equal(t, uint64(1), results[0].UserID)
	require.Equal(t, uint64(2), results[1].UserID)
	require.Equal(t, 0, results[0].Attendance)

	paidTrue := true
	results, total, err = repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA, Paid: &paidTrue})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, m1.ID, results[0].ID)

	limit := 1
	offset := 1
	results, total, err = repo.List(&models.ListMembershipsFilter{
		SemesterID: &semesterA,
		Pagination: models.Pagination{Limit: &limit, Offset: &offset},
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, results, 1)
	require.Equal(t, uint64(2), results[0].UserID)

	name := "bob"
	results, total, err = repo.List(&models.ListMembershipsFilter{SemesterID: &semesterA, Name: &name})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, uint64(2), results[0].UserID)

	otherSemester := uuid.New()
	results, total, err = repo.List(&models.ListMembershipsFilter{SemesterID: &otherSemester})
	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, results)
}

func TestMembershipRepository_Update(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	membership := &models.Membership{UserID: 1, SemesterID: uuid.New(), Paid: true, Discounted: true}
	require.NoError(t, repo.Create(membership))

	update := &models.Membership{ID: membership.ID, Paid: false, Discounted: false}
	require.NoError(t, repo.Update(update))

	found, err := repo.FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, found.Paid)
	require.False(t, found.Discounted)
}

func TestMembershipRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	err := repo.Update(&models.Membership{ID: uuid.New()})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_Delete(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	semesterA := uuid.New()
	semesterB := uuid.New()

	membership := &models.Membership{UserID: 1, SemesterID: semesterA}
	require.NoError(t, repo.Create(membership))

	err := repo.Delete(membership.ID, semesterB)
	require.ErrorIs(t, err, store.ErrNotFound)

	require.NoError(t, repo.Delete(membership.ID, semesterA))

	_, err = repo.FindByID(membership.ID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestMembershipRepository_Clone_Isolation(t *testing.T) {
	t.Parallel()

	repo := newMembershipRepository()

	membership := &models.Membership{UserID: 1, SemesterID: uuid.New(), Paid: false}
	require.NoError(t, repo.Create(membership))

	clone := repo.clone()

	// Mutating the clone must not affect the original.
	require.NoError(t, clone.Update(&models.Membership{ID: membership.ID, Paid: true}))

	original, err := repo.FindByID(membership.ID)
	require.NoError(t, err)
	require.False(t, original.Paid)

	cloned, err := clone.FindByID(membership.ID)
	require.NoError(t, err)
	require.True(t, cloned.Paid)

	// Creating a new membership on the original must not appear in the clone.
	require.NoError(t, repo.Create(&models.Membership{UserID: 2, SemesterID: uuid.New()}))

	_, cloneTotal, err := clone.List(&models.ListMembershipsFilter{})
	require.NoError(t, err)
	require.EqualValues(t, 1, cloneTotal)

	_, originalTotal, err := repo.List(&models.ListMembershipsFilter{})
	require.NoError(t, err)
	require.EqualValues(t, 2, originalTotal)
}
