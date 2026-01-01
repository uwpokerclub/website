package services

import (
	e "api/internal/errors"
	"api/internal/database"
	"api/internal/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginService(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "CreateLogin",
			test: CreateLoginTest,
		},
		{
			name: "ListLogins_Empty",
			test: ListLoginsEmptyTest,
		},
		{
			name: "ListLogins_Multiple",
			test: ListLoginsMultipleTest,
		},
		{
			name: "ListLogins_WithLinkedMember",
			test: ListLoginsWithLinkedMemberTest,
		},
		{
			name: "GetLogin",
			test: GetLoginTest,
		},
		{
			name: "GetLogin_WithLinkedMember",
			test: GetLoginWithLinkedMemberTest,
		},
		{
			name: "GetLogin_NotFound",
			test: GetLoginNotFoundTest,
		},
		{
			name: "DeleteLogin",
			test: DeleteLoginTest,
		},
		{
			name: "DeleteLogin_NotFound",
			test: DeleteLoginNotFoundTest,
		},
		{
			name: "ChangePassword",
			test: ChangePasswordTest,
		},
		{
			name: "ChangePassword_NotFound",
			test: ChangePasswordNotFoundTest,
		},
		{
			name: "CreateLoginFromRequest",
			test: CreateLoginFromRequestTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func CreateLoginTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	err = svc.CreateLogin("testuser", "password123", "executive")
	assert.NoError(t, err)

	// Verify login was created
	var login models.Login
	err = db.Where("username = ?", "testuser").First(&login).Error
	assert.NoError(t, err)
	assert.Equal(t, "testuser", login.Username)
	assert.Equal(t, "executive", login.Role)

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("password123"))
	assert.NoError(t, err)
}

func ListLoginsEmptyTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// List logins when none exist
	logins, err := svc.ListLogins()
	assert.NoError(t, err)
	assert.Empty(t, logins)
}

func ListLoginsMultipleTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Create test logins
	err = svc.CreateLogin("alice", "password123", "executive")
	assert.NoError(t, err)
	err = svc.CreateLogin("bob", "password456", "president")
	assert.NoError(t, err)
	err = svc.CreateLogin("charlie", "password789", "webmaster")
	assert.NoError(t, err)

	// List logins
	logins, err := svc.ListLogins()
	assert.NoError(t, err)
	assert.Len(t, logins, 3)

	// Verify ordering (alphabetical by username)
	assert.Equal(t, "alice", logins[0].Username)
	assert.Equal(t, "bob", logins[1].Username)
	assert.Equal(t, "charlie", logins[2].Username)

	// Verify roles
	assert.Equal(t, "executive", logins[0].Role)
	assert.Equal(t, "president", logins[1].Role)
	assert.Equal(t, "webmaster", logins[2].Role)

	// Verify no linked members (since we didn't create users)
	assert.Nil(t, logins[0].LinkedMember)
	assert.Nil(t, logins[1].LinkedMember)
	assert.Nil(t, logins[2].LinkedMember)
}

func ListLoginsWithLinkedMemberTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)
	userSvc := NewUserService(db)

	// Create a user with QuestID
	user, err := userSvc.CreateUser(&models.CreateUserRequest{
		ID:        12345678,
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		Faculty:   "Math",
		QuestID:   "asmith",
	})
	assert.NoError(t, err)

	// Create login matching QuestID
	err = svc.CreateLogin("asmith", "password123", "executive")
	assert.NoError(t, err)

	// Create login without matching user
	err = svc.CreateLogin("webmaster", "password456", "webmaster")
	assert.NoError(t, err)

	// List logins
	logins, err := svc.ListLogins()
	assert.NoError(t, err)
	assert.Len(t, logins, 2)

	// Find alice login (alphabetically first)
	var aliceLogin *models.LoginWithMember
	var webmasterLogin *models.LoginWithMember
	for i := range logins {
		if logins[i].Username == "asmith" {
			aliceLogin = &logins[i]
		}
		if logins[i].Username == "webmaster" {
			webmasterLogin = &logins[i]
		}
	}

	// Verify alice has linked member
	assert.NotNil(t, aliceLogin)
	assert.NotNil(t, aliceLogin.LinkedMember)
	assert.Equal(t, user.ID, aliceLogin.LinkedMember.ID)
	assert.Equal(t, "Alice", aliceLogin.LinkedMember.FirstName)
	assert.Equal(t, "Smith", aliceLogin.LinkedMember.LastName)

	// Verify webmaster has no linked member
	assert.NotNil(t, webmasterLogin)
	assert.Nil(t, webmasterLogin.LinkedMember)
}

func GetLoginTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Create test login
	err = svc.CreateLogin("alice", "password123", "executive")
	assert.NoError(t, err)

	// Get login
	login, err := svc.GetLogin("alice")
	assert.NoError(t, err)
	assert.Equal(t, "alice", login.Username)
	assert.Equal(t, "executive", login.Role)
	assert.Nil(t, login.LinkedMember)
}

func GetLoginWithLinkedMemberTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)
	userSvc := NewUserService(db)

	// Create user
	user, err := userSvc.CreateUser(&models.CreateUserRequest{
		ID:        12345678,
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		Faculty:   "Math",
		QuestID:   "asmith",
	})
	assert.NoError(t, err)

	// Create matching login
	err = svc.CreateLogin("asmith", "password123", "executive")
	assert.NoError(t, err)

	// Get login
	login, err := svc.GetLogin("asmith")
	assert.NoError(t, err)
	assert.NotNil(t, login.LinkedMember)
	assert.Equal(t, user.ID, login.LinkedMember.ID)
	assert.Equal(t, "Alice", login.LinkedMember.FirstName)
	assert.Equal(t, "Smith", login.LinkedMember.LastName)
}

func GetLoginNotFoundTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Try to get non-existent login
	_, err = svc.GetLogin("nonexistent")
	assert.Error(t, err)

	// Verify it's a NotFound error
	apiErr, ok := err.(e.APIErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.Code)
}

func DeleteLoginTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Create test login
	err = svc.CreateLogin("alice", "password123", "executive")
	assert.NoError(t, err)

	// Verify login exists
	_, err = svc.GetLogin("alice")
	assert.NoError(t, err)

	// Delete login
	err = svc.DeleteLogin("alice")
	assert.NoError(t, err)

	// Verify login is gone
	_, err = svc.GetLogin("alice")
	assert.Error(t, err)

	apiErr, ok := err.(e.APIErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.Code)
}

func DeleteLoginNotFoundTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Try to delete non-existent login
	err = svc.DeleteLogin("nonexistent")
	assert.Error(t, err)

	// Verify it's a NotFound error
	apiErr, ok := err.(e.APIErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.Code)
}

func ChangePasswordTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Create test login
	err = svc.CreateLogin("alice", "oldpassword", "executive")
	assert.NoError(t, err)

	// Change password
	err = svc.ChangePassword("alice", "newpassword123")
	assert.NoError(t, err)

	// Verify password was changed by checking the hash
	var login models.Login
	err = db.Where("username = ?", "alice").First(&login).Error
	assert.NoError(t, err)

	// Verify the new password works
	err = bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("newpassword123"))
	assert.NoError(t, err)

	// Verify the old password doesn't work
	err = bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("oldpassword"))
	assert.Error(t, err)
}

func ChangePasswordNotFoundTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Try to change password for non-existent login
	err = svc.ChangePassword("nonexistent", "newpassword")
	assert.Error(t, err)

	// Verify it's a NotFound error
	apiErr, ok := err.(e.APIErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.Code)
}

func CreateLoginFromRequestTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	svc := NewLoginService(db)

	// Create login from request
	req := &models.CreateLoginRequest{
		Username: "alice",
		Password: "password123",
		Role:     "executive",
	}
	err = svc.CreateLoginFromRequest(req)
	assert.NoError(t, err)

	// Verify login was created
	login, err := svc.GetLogin("alice")
	assert.NoError(t, err)
	assert.Equal(t, "alice", login.Username)
	assert.Equal(t, "executive", login.Role)

	// Verify password was hashed correctly
	var rawLogin models.Login
	err = db.Where("username = ?", "alice").First(&rawLogin).Error
	assert.NoError(t, err)
	err = bcrypt.CompareHashAndPassword([]byte(rawLogin.Password), []byte("password123"))
	assert.NoError(t, err)
}
