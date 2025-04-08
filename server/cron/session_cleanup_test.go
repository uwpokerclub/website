package cron

import (
	"api/internal/database"
	"api/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func createTestLogin(db *gorm.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	login := models.Login{
		Username: username,
		Password: string(hash),
	}

	res := db.Create(&login)

	return res.Error
}

func createTestSession(db *gorm.DB, username, password string, start time.Time) (*models.Session, error) {
	// Create test user
	err := createTestLogin(db, username, password)
	if err != nil {
		return nil, err
	}

	// Create test session
	session := models.Session{
		ID:        uuid.New(),
		Username:  username,
		StartedAt: start,
		ExpiresAt: start.Add(time.Hour * 8),
	}
	res := db.Create(&session)

	return &session, res.Error
}

func TestSessionCleanup(t *testing.T) {
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
	defer wipeDB()

	now := time.Now().UTC()

	session1 := now.Add(time.Hour * -1)
	session2 := now.Add(time.Hour * -4).Add(time.Minute * -30)
	session3 := now.Add(time.Hour * -7)
	session4 := now.Add(time.Hour * -9)
	session5 := now.Add(time.Hour * -12)

	s1, err := createTestSession(db, "user1", "pass", session1)
	assert.NoError(t, err, "Creating test session 1 should not error")
	s2, err := createTestSession(db, "user2", "pass", session2)
	assert.NoError(t, err, "Creating test session 2 should not error")
	s3, err := createTestSession(db, "user3", "pass", session3)
	assert.NoError(t, err, "Creating test session 3 should not error")
	_, err = createTestSession(db, "user4", "pass", session4)
	assert.NoError(t, err, "Creating test session 4 should not error")
	_, err = createTestSession(db, "user5", "pass", session5)
	assert.NoError(t, err, "Creating test session 5 should not error")

	// List of non-expired session ids
	expectedSessionIds := []string{s1.Username, s2.Username, s3.Username}

	// Run cron job
	SessionCleanup(true)()

	// Check if only the expired sessions (s4, s5) were deleted from the database
	var sessionIds []string
	res := db.Table("sessions").Select("username").Find(&sessionIds)
	assert.NoError(t, res.Error, "Should be no database error")
	assert.Equal(t, len(expectedSessionIds), len(sessionIds), "Session count should match")
	assert.ElementsMatch(t, expectedSessionIds, sessionIds, "Expected session ids (list A) should match existing session ids (list B)")
}
