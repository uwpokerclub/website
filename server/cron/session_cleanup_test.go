package cron

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"api/internal/models"
	"api/internal/testutils"
)

func createTestSession(db *gorm.DB, username string, start time.Time) (*models.Session, error) {
	login := models.Login{
		Username: username,
		Password: "not_used",
		Role:     "executive",
	}
	if err := db.Create(&login).Error; err != nil {
		return nil, err
	}

	session := models.Session{
		ID:        uuid.New(),
		Username:  username,
		StartedAt: start,
		ExpiresAt: start.Add(time.Hour * 8),
		Role:      "executive",
	}
	if err := db.Create(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func TestSessionCleanup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	now := time.Now().UTC()

	// Sessions started 1, 4.5, and 7 hours ago (expire in 7, 3.5, and 1 hours — still valid)
	s1, err := createTestSession(db, "user1", now.Add(time.Hour*-1))
	require.NoError(t, err)
	s2, err := createTestSession(db, "user2", now.Add(time.Hour*-4+time.Minute*-30))
	require.NoError(t, err)
	s3, err := createTestSession(db, "user3", now.Add(time.Hour*-7))
	require.NoError(t, err)

	// Sessions started 9 and 12 hours ago (expired 1 and 4 hours ago)
	_, err = createTestSession(db, "user4", now.Add(time.Hour*-9))
	require.NoError(t, err)
	_, err = createTestSession(db, "user5", now.Add(time.Hour*-12))
	require.NoError(t, err)

	expectedIDs := []uuid.UUID{s1.ID, s2.ID, s3.ID}

	// Run cron job
	SessionCleanup(db)()

	// Check that only the expired sessions (s4, s5) were deleted
	var remainingIDs []uuid.UUID
	res := db.Table("sessions").Select("id").Find(&remainingIDs)
	assert.NoError(t, res.Error)
	assert.Equal(t, len(expectedIDs), len(remainingIDs))
	assert.ElementsMatch(t, expectedIDs, remainingIDs)
}
