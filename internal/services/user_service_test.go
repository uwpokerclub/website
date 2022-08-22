package services

import (
	"api/internal/database"
	"api/internal/models"
	"errors"
	"reflect"
	"testing"

	"gorm.io/gorm"
)

func TestUserService(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "CreateUser",
			test: CreateUserTest(),
		},
		{
			name: "ListUsers",
			test: ListUsersTest(),
		},
		{
			name: "GetUser",
			test: GetUserTest(),
		},
		{
			name: "GetUser_NotFound",
			test: GetUserNotFoundTest(),
		},
		{
			name: "UpdateUser",
			test: UpdateUserTest(),
		},
		{
			name: "DeleteUser",
			test: DeleteUserTest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func CreateUserTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		us := NewUserService(db)

		req := &models.CreateUserRequest{
			ID:        1,
			FirstName: "adam",
			LastName:  "mahood",
			Email:     "adam@gmail.com",
			Faculty:   "Math",
			QuestID:   "asmahood",
		}

		res, err := us.CreateUser(req)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
			return
		}

		exp := &models.User{
			ID:        req.ID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Faculty:   req.Faculty,
			QuestID:   req.QuestID,
			// Workaround since we can't ensure the timestamps are perfect
			CreatedAt: res.CreatedAt,
		}

		// Ensure correct user is created
		if !reflect.DeepEqual(exp, res) {
			t.Errorf("Created user does not match the expected return. Expected: %v, Got: %v", exp, res)
		}
	}
}

func ListUsersTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		// Create existing users
		user1 := models.User{
			ID:        1,
			FirstName: "adam",
			LastName:  "mahood",
			Email:     "adam@gmail.com",
			Faculty:   "Math",
			QuestID:   "asmahood",
		}
		res := db.Create(&user1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing users: %v", res.Error)
		}

		user2 := models.User{
			ID:        2,
			FirstName: "john",
			LastName:  "doe",
			Email:     "john@gmail.com",
			Faculty:   "Science",
			QuestID:   "jdoe",
		}
		res = db.Create(&user2)
		if res.Error != nil {
			t.Fatalf("Error when creating existing users: %v", res.Error)
		}

		user3 := models.User{
			ID:        3,
			FirstName: "jane",
			LastName:  "doe",
			Email:     "jane@gmail.com",
			Faculty:   "Arts",
			QuestID:   "jadoe",
		}
		res = db.Create(&user3)
		if res.Error != nil {
			t.Fatalf("Error when creating existing users: %v", res.Error)
		}

		us := NewUserService(db)

		users, err := us.ListUsers()
		// Should not return an error
		if err != nil {
			t.Errorf("Unexpected error occurred in ListUsers: %v", err)
			return
		}

		// Should return 3 elements, in order of newest to oldest
		if len(users) != 3 {
			t.Errorf("Expected 3 elements in array, got %d", len(users))
			return
		}

		// Workaround since timestamps cannot be deep equality checked
		user1.CreatedAt = users[0].CreatedAt
		user2.CreatedAt = users[1].CreatedAt
		user3.CreatedAt = users[2].CreatedAt

		// Array should have each user
		expUsers := []models.User{user1, user2, user3}
		if !reflect.DeepEqual(expUsers, users) {
			t.Errorf("Expected return and actual return does not match. Expected: %v, Got: %v", expUsers, users)
			return
		}
	}
}

func GetUserTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		// Create existing user
		user1 := models.User{
			ID:        1,
			FirstName: "adam",
			LastName:  "mahood",
			Email:     "adam@gmail.com",
			Faculty:   "Math",
			QuestID:   "asmahood",
		}
		res := db.Create(&user1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing user: %v", res.Error)
		}

		us := NewUserService(db)

		user, err := us.GetUser(1)
		// Should not return an error
		if err != nil {
			t.Errorf("UserService.GetUser() errored: %v", err)
			return
		}

		// Workaround since timestamps cannot be deep equality checked
		user1.CreatedAt = user.CreatedAt

		// Expected user and actual user should be equal
		if !reflect.DeepEqual(user1, *user) {
			t.Errorf("UserService.GetUser() = %v, wanted: %v", user, user1)
			return
		}
	}
}

func GetUserNotFoundTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		us := NewUserService(db)

		_, err = us.GetUser(1)
		// Should return an error
		if err == nil {
			t.Errorf("UserService.GetUser() did not error, expected error")
			return
		}
	}
}

func UpdateUserTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		// Create existing user
		user1 := models.User{
			ID:        1,
			FirstName: "adam",
			LastName:  "mahood",
			Email:     "adam@gmail.com",
			Faculty:   "Math",
			QuestID:   "asmahood",
		}
		res := db.Create(&user1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing user: %v", res.Error)
		}

		us := NewUserService(db)

		req := models.UpdateUserRequest{
			Email: "adam.updated@gmail.com",
		}

		updatedUser, err := us.UpdateUser(user1.ID, &req)
		if err != nil {
			t.Errorf("UserService.UpdateUser() errored: %v", err)
			return
		}

		exp := models.User{
			ID:        1,
			FirstName: "adam",
			LastName:  "mahood",
			Email:     "adam.updated@gmail.com",
			Faculty:   "Math",
			QuestID:   "asmahood",
			CreatedAt: updatedUser.CreatedAt,
		}

		// Expect only the email to be updated
		if user1.Email == updatedUser.Email {
			t.Errorf("UserService.UpdatedUser().Email not updated")
			return
		}

		// Expect updated user and expected to match
		if !reflect.DeepEqual(exp, *updatedUser) {
			t.Errorf("UserService.UpdateUser() = %v, wanted: %v", *updatedUser, exp)
			return
		}
	}
}

func DeleteUserTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		// Create existing user
		user1 := models.User{
			ID:        1,
			FirstName: "adam",
			LastName:  "mahood",
			Email:     "adam@gmail.com",
			Faculty:   "Math",
			QuestID:   "asmahood",
		}
		res := db.Create(&user1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing user: %v", res.Error)
		}

		us := NewUserService(db)

		err = us.DeleteUser(user1.ID)
		if err != nil {
			t.Errorf("UserService.DeleteUser() errored: %v", err)
			return
		}

		// Ensure user was deleted from the db
		foundUser := models.User{ID: user1.ID}

		res = db.First(&foundUser)
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			t.Errorf("UserService.DeleteUser() user was not deleted from the DB: %v", res.Error)
			return
		}
	}
}
