package authentication_test

import (
	"api/internal/authentication"
	"api/internal/models"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateTestLogin(db *gorm.DB, username, password, status string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	login := models.Login{
		Username: username,
		Password: string(hash),
		Status:   status,
	}

	res := db.Create(&login)

	return res.Error
}

func TestCredentialsService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()

	resetDB := func() {
		require.NoError(t, container.ResetDatabase(ctx))
	}

	credSvc := authentication.NewCredentialService(postgres.NewStore(db))

	t.Run("Validate_IncorrectUsername", func(t *testing.T) {
		t.Cleanup(resetDB)

		// Create test user
		err := CreateTestLogin(db, "testuser", "password", models.LoginStatusActive)
		assert.NoError(t, err)

		valid, role, err := credSvc.Validate("nouser", "password")
		assert.NoError(t, err)
		assert.False(t, valid)
		assert.Empty(t, role)
	})

	t.Run("Validate_IncorrectPassword", func(t *testing.T) {
		t.Cleanup(resetDB)

		// Create test user
		err := CreateTestLogin(db, "testuser", "password", models.LoginStatusActive)
		assert.NoError(t, err)

		valid, role, err := credSvc.Validate("testuser", "wrongpassword")
		assert.NoError(t, err)
		assert.False(t, valid)
		assert.Empty(t, role)
	})

	t.Run("Validate_CorrectCredentials", func(t *testing.T) {
		t.Cleanup(resetDB)

		// Create test user
		err := CreateTestLogin(db, "testuser", "password", models.LoginStatusActive)
		assert.NoError(t, err)

		valid, role, err := credSvc.Validate("testuser", "password")
		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Equal(t, "executive", role)
	})

	for _, status := range []string{models.LoginStatusDisabled, models.LoginStatusPendingActivation} {
		t.Run("Validate_"+status+"Login", func(t *testing.T) {
			t.Cleanup(resetDB)

			require.NoError(t, CreateTestLogin(db, "testuser", "password", status))

			valid, role, err := credSvc.Validate("testuser", "password")
			require.NoError(t, err)
			assert.False(t, valid)
			assert.Empty(t, role)
		})
	}
}
