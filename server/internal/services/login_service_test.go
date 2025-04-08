package services

import (
	"api/internal/database"
	"api/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginService(t *testing.T) {
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

	loginService := NewLoginService(db)

	t.Run("CreateLogin", func(t *testing.T) {
		t.Cleanup(wipeDB)

		err := loginService.CreateLogin("testuser", "testpassword", "president")
		if assert.NoError(t, err) {
			var login models.Login
			err := db.First(&login, "username = ?", "testuser").Error
			if err != nil {
				t.Fatal(err.Error())
			}
			assert.Equal(t, "testuser", login.Username)
			assert.NotEmpty(t, login.Password)
			assert.Equal(t, "president", login.Role)
		}
	})
}
