package inmemory

import (
	"testing"
	"time"

	"api/internal/models"
	"api/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestParticipant(membershipID uuid.UUID, eventID int32) *models.Participant {
	return &models.Participant{
		MembershipID: &membershipID,
		EventID:      eventID,
	}
}

func TestEntryRepository_Create(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	participant := newTestParticipant(uuid.New(), 1)
	require.NoError(t, repo.Create(participant))
	require.NotZero(t, participant.ID)
}

func TestEntryRepository_Create_DuplicateMembershipEvent(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()

	require.NoError(t, repo.Create(newTestParticipant(membershipID, 1)))

	err := repo.Create(newTestParticipant(membershipID, 1))
	require.Error(t, err)
}

func TestEntryRepository_FindByID(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	participant := newTestParticipant(uuid.New(), 1)
	require.NoError(t, repo.Create(participant))

	found, err := repo.FindByID(participant.ID)
	require.NoError(t, err)
	require.Equal(t, participant.ID, found.ID)

	_, err = repo.FindByID(99999)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_FindByMembershipAndEventID(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()

	created := newTestParticipant(membershipID, 1)
	require.NoError(t, repo.Create(created))

	found, err := repo.FindByMembershipAndEventID(membershipID, 1)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)

	_, err = repo.FindByMembershipAndEventID(membershipID, 2)
	require.ErrorIs(t, err, store.ErrNotFound)

	_, err = repo.FindByMembershipAndEventID(uuid.New(), 1)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_List(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	stillIn := newTestParticipant(uuid.New(), 1)
	require.NoError(t, repo.Create(stillIn))

	signedOutEarlier := newTestParticipant(uuid.New(), 1)
	require.NoError(t, repo.Create(signedOutEarlier))
	earlier := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	require.NoError(t, repo.Update(signedOutEarlier, map[string]any{"signed_out_at": &earlier}))

	signedOutLater := newTestParticipant(uuid.New(), 1)
	require.NoError(t, repo.Create(signedOutLater))
	later := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, repo.Update(signedOutLater, map[string]any{"signed_out_at": &later}))

	// Different event -- must be excluded.
	require.NoError(t, repo.Create(newTestParticipant(uuid.New(), 2)))

	participants, total, err := repo.List(&models.ListParticipantsFilter{EventID: 1})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, participants, 3)

	require.Equal(t, stillIn.ID, participants[0].ID)
	require.Equal(t, signedOutLater.ID, participants[1].ID)
	require.Equal(t, signedOutEarlier.ID, participants[2].ID)
}

func TestEntryRepository_Update_SignOutThenSignIn(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	participant := newTestParticipant(uuid.New(), 1)
	require.NoError(t, repo.Create(participant))
	require.Nil(t, participant.SignedOutAt)

	now := time.Now().UTC()
	require.NoError(t, repo.Update(participant, map[string]any{"signed_out_at": &now}))
	require.NotNil(t, participant.SignedOutAt)
	require.Equal(t, now, *participant.SignedOutAt)

	found, err := repo.FindByID(participant.ID)
	require.NoError(t, err)
	require.NotNil(t, found.SignedOutAt)

	require.NoError(t, repo.Update(participant, map[string]any{"signed_out_at": nil}))
	require.Nil(t, participant.SignedOutAt)

	found, err = repo.FindByID(participant.ID)
	require.NoError(t, err)
	require.Nil(t, found.SignedOutAt)
}

func TestEntryRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	err := repo.Update(&models.Participant{ID: 99999}, map[string]any{"signed_out_at": nil})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_Delete(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()

	require.NoError(t, repo.Create(newTestParticipant(membershipID, 1)))
	require.NoError(t, repo.Delete(membershipID, 1))

	_, err := repo.FindByMembershipAndEventID(membershipID, 1)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	err := repo.Delete(uuid.New(), 1)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestEntryRepository_Clone_Isolation(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()

	participant := newTestParticipant(membershipID, 1)
	require.NoError(t, repo.Create(participant))

	clone := repo.clone()

	// Mutating the clone must not affect the original.
	now := time.Now().UTC()
	require.NoError(t, clone.Update(&models.Participant{ID: participant.ID}, map[string]any{"signed_out_at": &now}))

	original, err := repo.FindByID(participant.ID)
	require.NoError(t, err)
	require.Nil(t, original.SignedOutAt)

	cloned, err := clone.FindByID(participant.ID)
	require.NoError(t, err)
	require.NotNil(t, cloned.SignedOutAt)

	// Creating a new entry on the original must not appear in the clone.
	require.NoError(t, repo.Create(newTestParticipant(uuid.New(), 1)))

	_, cloneTotal, err := clone.List(&models.ListParticipantsFilter{EventID: 1})
	require.NoError(t, err)
	require.EqualValues(t, 1, cloneTotal)

	_, originalTotal, err := repo.List(&models.ListParticipantsFilter{EventID: 1})
	require.NoError(t, err)
	require.EqualValues(t, 2, originalTotal)
}

func TestEntryRepository_List_Search(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()

	membershipID := uuid.New()
	participant := &models.Participant{
		MembershipID: &membershipID,
		EventID:      1,
		Membership: &models.Membership{
			ID:   membershipID,
			User: &models.User{FirstName: "Katherine", LastName: "Johnson"},
		},
	}
	require.NoError(t, repo.Create(participant))

	results, total, err := repo.List(&models.ListParticipantsFilter{EventID: 1, Search: "katherine"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, results, 1)

	results, total, err = repo.List(&models.ListParticipantsFilter{EventID: 1, Search: "nobody"})
	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, results)
}

func TestEntryRepository_SignOutAllUnsigned(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()
	require.NoError(t, repo.Create(&models.Participant{MembershipID: &membershipID, EventID: 1}))

	now := time.Now().UTC()
	require.NoError(t, repo.SignOutAllUnsigned(1, now))

	found, err := repo.FindByMembershipAndEventID(membershipID, 1)
	require.NoError(t, err)
	require.NotNil(t, found.SignedOutAt)
	require.Equal(t, now, *found.SignedOutAt)
}

func TestEntryRepository_BatchUpdatePlacements(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()
	participant := &models.Participant{MembershipID: &membershipID, EventID: 1}
	require.NoError(t, repo.Create(participant))

	require.NoError(t, repo.BatchUpdatePlacements(map[int32]uint16{participant.ID: 3}))

	found, err := repo.FindByID(participant.ID)
	require.NoError(t, err)
	require.EqualValues(t, 3, found.Placement)
}

func TestEntryRepository_CountByMembershipID(t *testing.T) {
	t.Parallel()

	repo := newEntryRepository()
	membershipID := uuid.New()
	otherMembershipID := uuid.New()

	require.NoError(t, repo.Create(newTestParticipant(membershipID, 1)))
	require.NoError(t, repo.Create(newTestParticipant(membershipID, 2)))
	require.NoError(t, repo.Create(newTestParticipant(otherMembershipID, 1)))

	count, err := repo.CountByMembershipID(membershipID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	count, err = repo.CountByMembershipID(uuid.New())
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}
