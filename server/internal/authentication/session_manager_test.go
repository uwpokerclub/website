package authentication_test

import (
	"api/internal/authentication"
	"api/internal/models"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func CreateTestSession(db *gorm.DB, username, password string, start time.Time) (*models.Session, error) {
	// Create test user
	err := CreateTestLogin(db, "testuser", "password")
	if err != nil {
		return nil, err
	}

	// Create test session
	session := models.Session{
		ID:        uuid.New(),
		Username:  "testuser",
		Role:      "executive",
		StartedAt: start,
		ExpiresAt: start.Add(time.Hour * 8),
	}
	res := db.Create(&session)

	return &session, res.Error
}

func TestSessionManager(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()

	resetDB := func() {
		require.NoError(t, container.ResetDatabase(ctx))
	}

	sessManager := authentication.NewSessionManager(postgres.NewStore(db))
	t.Run("Create_NoAssociatedLogin", func(t *testing.T) {
		t.Cleanup(resetDB)

		_, err := sessManager.Create("testuser", "executive")
		assert.Error(t, err)
	})
	t.Run("Create", func(t *testing.T) {
		t.Cleanup(resetDB)

		err := CreateTestLogin(db, "testuser", "password")
		assert.NoError(t, err)

		id, err := sessManager.Create("testuser", "executive")
		assert.NoError(t, err)
		assert.NoError(t, uuid.Validate(id.String()))

		session := models.Session{ID: id}
		res := db.First(&session)
		assert.NoError(t, res.Error)
		assert.Equal(t, "testuser", session.Username)
		assert.Equal(t, "executive", session.Role)
		assert.WithinDuration(t, session.StartedAt, session.ExpiresAt, time.Hour*8)
	})

	t.Run("Invalidate__NoAssociatedSession", func(t *testing.T) {
		t.Cleanup(resetDB)

		err := sessManager.Invalidate(uuid.New())
		assert.NoError(t, err)
	})

	t.Run("Invalidate__ExistingSession", func(t *testing.T) {
		t.Cleanup(resetDB)

		// Create test session
		session, err := CreateTestSession(db, "testuser", "password", time.Now())
		assert.NoError(t, err)

		err = sessManager.Invalidate(session.ID)
		assert.NoError(t, err)

		// Check to ensure session was deleted
		foundSession := models.Session{
			ID: session.ID,
		}
		res := db.First(&foundSession)
		assert.Error(t, res.Error)
		assert.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)
	})

	t.Run("Authenticate__NoSession", func(t *testing.T) {
		t.Cleanup(resetDB)

		_, err := sessManager.Authenticate(uuid.New())
		assert.ErrorIs(t, err, authentication.ErrSessionNotFound)
	})

	t.Run("Authenticate__SessionExpired", func(t *testing.T) {
		t.Cleanup(resetDB)

		session, err := CreateTestSession(db, "testuser", "password", time.Now().Add(time.Hour*-9))
		assert.NoError(t, err)

		_, err = sessManager.Authenticate(session.ID)
		assert.ErrorIs(t, err, authentication.ErrSessionExpired)

		// Ensure session was deleted
		foundSession := models.Session{
			ID: session.ID,
		}
		res := db.First(&foundSession)
		assert.Error(t, res.Error)
		assert.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)
	})

	t.Run("Authenticate__ValidSession", func(t *testing.T) {
		t.Cleanup(resetDB)

		session, err := CreateTestSession(db, "testuser", "password", time.Now())
		assert.NoError(t, err)

		foundSession, err := sessManager.Authenticate(session.ID)
		assert.NoError(t, err)
		assert.Equal(t, session.Username, foundSession.Username)
	})
}
