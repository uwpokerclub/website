package authentication

import (
	"api/internal/database"
	"api/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
		StartedAt: start,
		ExpiresAt: start.Add(time.Hour * 8),
	}
	res := db.Create(&session)

	return &session, res.Error
}

func TestSessionManager(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sqlDB.Close()

	wipeDB := func() {
		err := database.WipeDB(db)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	sessManager := NewSessionManager(db)
	t.Run("Create_NoAssociatedLogin", func(t *testing.T) {
		t.Cleanup(wipeDB)

		_, err := sessManager.Create("testuser")
		assert.Error(t, err)
	})
	t.Run("Create", func(t *testing.T) {
		t.Cleanup(wipeDB)

		err := CreateTestLogin(db, "testuser", "password")
		assert.NoError(t, err)

		id, err := sessManager.Create("testuser")
		assert.NoError(t, err)
		assert.NoError(t, uuid.Validate(id.String()))

		session := models.Session{ID: id}
		res := db.First(&session)
		assert.NoError(t, res.Error)
		assert.Equal(t, "testuser", session.Username)
		assert.WithinDuration(t, session.StartedAt, session.ExpiresAt, time.Hour*8)
	})

	t.Run("Invalidate__NoAssociatedSession", func(t *testing.T) {
		t.Cleanup(wipeDB)

		err := sessManager.Invalidate(uuid.New())
		assert.NoError(t, err)
	})

	t.Run("Invalidate__ExistingSession", func(t *testing.T) {
		t.Cleanup(wipeDB)

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
		t.Cleanup(wipeDB)

		err := sessManager.Authenticate(uuid.New())
		assert.Error(t, err)
	})

	t.Run("Authenticate__SessionExpired", func(t *testing.T) {
		t.Cleanup(wipeDB)

		session, err := CreateTestSession(db, "testuser", "password", time.Now().Add(time.Hour*-9))
		assert.NoError(t, err)

		err = sessManager.Authenticate(session.ID)
		assert.Error(t, err)

		// Ensure session was deleted
		foundSession := models.Session{
			ID: session.ID,
		}
		res := db.First(&foundSession)
		assert.Error(t, res.Error)
		assert.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)
	})

	t.Run("Authenticate__ValidSession", func(t *testing.T) {
		t.Cleanup(wipeDB)

		session, err := CreateTestSession(db, "testuser", "password", time.Now())
		assert.NoError(t, err)

		err = sessManager.Authenticate(session.ID)
		assert.NoError(t, err)
	})

	t.Run("Get__NoSession", func(t *testing.T) {
		t.Cleanup(wipeDB)
		session, err := sessManager.Get(uuid.New())
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Get__ValidSession", func(t *testing.T) {
		t.Cleanup(wipeDB)
		session, err := CreateTestSession(db, "testuser", "password", time.Now())
		assert.NoError(t, err)

		foundSession, err := sessManager.Get(session.ID)
		assert.NoError(t, err)
		assert.Equal(t, session.Username, foundSession.Username)
	})
}
