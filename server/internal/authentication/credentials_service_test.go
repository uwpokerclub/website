package authentication

import (
	"api/internal/database"
	"api/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateTestLogin(db *gorm.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	login := models.Login{
		Username: "testuser",
		Password: string(hash),
	}

	res := db.Create(&login)

	return res.Error
}

func TestCredentialsService(t *testing.T) {
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

	credSvc := NewCredentialService(db)

	t.Run("Validate_IncorrectUsername", func(t *testing.T) {
		t.Cleanup(wipeDB)

		// Create test user
		err := CreateTestLogin(db, "testuser", "password")
		assert.NoError(t, err)

		valid, err := credSvc.Validate("nouser", "password")
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Validate_IncorrectPassword", func(t *testing.T) {
		t.Cleanup(wipeDB)

		// Create test user
		err := CreateTestLogin(db, "testuser", "password")
		assert.NoError(t, err)

		valid, err := credSvc.Validate("testuser", "wrongpassword")
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Validate_CorrectCredentials", func(t *testing.T) {
		t.Cleanup(wipeDB)

		// Create test user
		err := CreateTestLogin(db, "testuser", "password")
		assert.NoError(t, err)

		valid, err := credSvc.Validate("testuser", "password")
		assert.NoError(t, err)
		assert.True(t, valid)
	})
}
