package authentication

import (
	"api/internal/database"
	"api/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
}
