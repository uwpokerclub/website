package authorization

import (
	"api/internal/database"
	"api/internal/errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestNewAuthorizationService(t *testing.T) {
	testCases := []struct {
		name          string
		mockDatabase  func(mock sqlmock.Sqlmock)
		username      string
		expectedRole  string
		expectError   bool
		expectedError error
	}{
		{
			name: "Can find user",
			mockDatabase: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username", "password", "role"}).AddRow("testuser", "password", "webmaster")

				mock.ExpectQuery("^SELECT \\* FROM \"logins\" WHERE \"logins\".\"username\" = \\$1$").WillReturnRows(rows)
			},
			username:      "testuser",
			expectedRole:  "webmaster",
			expectError:   false,
			expectedError: nil,
		},
		{
			name: "Cannot find user",
			mockDatabase: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"logins\" WHERE \"logins\".\"username\" = \\$1$").WillReturnError(gorm.ErrRecordNotFound)
			},
			username:      "testuser",
			expectedRole:  "",
			expectError:   true,
			expectedError: errors.Unauthorized("You do not have permission to perform this action."),
		},
		{
			name: "Database error",
			mockDatabase: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"logins\" WHERE \"logins\".\"username\" = \\$1$").WillReturnError(gorm.ErrInvalidField)
			},
			username:      "testuser",
			expectedRole:  "",
			expectError:   true,
			expectedError: errors.InternalServerError("An internal problem has occurred. Please try again later."),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			db, mock, err := database.NewMockDatabase()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}

			tC.mockDatabase(mock)

			authService, err := NewAuthorizationService(db, tC.username, nil)
			if tC.expectError {
				assert.Error(t, err)
				if tC.expectedError != nil {
					assert.ErrorIs(t, err, tC.expectedError)
				}
				assert.Nil(t, authService)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, authService)
				assert.Equal(t, tC.expectedRole, authService.role)
			}
		})
	}
}

type MockResourceAuthorizer struct {
	mock.Mock
}

func (m *MockResourceAuthorizer) IsAuthorized(role string, action string) bool {
	args := m.Called(role, action)
	return args.Bool(0)
}

func (m *MockResourceAuthorizer) GetPermissions(role string) map[string]any {
	args := m.Called(0)
	return args.Get(0).(map[string]any)
}

func TestAuthorizationService_IsAuthorized(t *testing.T) {
	sqlMock := func(mock sqlmock.Sqlmock) {
		rows := sqlmock.NewRows([]string{"username", "password", "role"}).AddRow("testuser", "password", "webmaster")
		mock.ExpectQuery("^SELECT \\* FROM \"logins\" WHERE \"logins\".\"username\" = \\$1$").WillReturnRows(rows)
	}

	testCases := []struct {
		name                   string
		mockDatabase           func(m sqlmock.Sqlmock)
		mockResourceAuthorizer func(m *MockResourceAuthorizer)
		resourceAuthorizers    ResourceAuthorizerMap
		action                 string
		expected               bool
	}{
		{
			name:                "Empty action",
			mockDatabase:        sqlMock,
			resourceAuthorizers: ResourceAuthorizerMap{},
			action:              "",
			expected:            false,
		},
		{
			name:                "Only resource supplied",
			mockDatabase:        sqlMock,
			resourceAuthorizers: ResourceAuthorizerMap{},
			action:              "resource",
			expected:            false,
		},
		{
			name:         "Unknown resource",
			mockDatabase: sqlMock,
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			action:   "other.create",
			expected: false,
		},
		{
			name:         "Resource Authorized",
			mockDatabase: sqlMock,
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("IsAuthorized", mock.Anything, mock.Anything).Return(true)
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			action:   "resource.create",
			expected: true,
		},
		{
			name:         "Resource Unauthorized",
			mockDatabase: sqlMock,
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("IsAuthorized", mock.Anything, mock.Anything).Return(false)
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			action:   "resource.create",
			expected: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			db, mock, err := database.NewMockDatabase()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}

			tC.mockDatabase(mock)
			if tC.mockResourceAuthorizer != nil {
				tC.mockResourceAuthorizer(tC.resourceAuthorizers["resource"].(*MockResourceAuthorizer))
			}

			authService, err := NewAuthorizationService(db, "testuser", tC.resourceAuthorizers)
			if err != nil {
				t.Fatalf("NewAuthorizationService constructor should not have errored!")
			}

			result := authService.IsAuthorized(tC.action)
			assert.Equal(t, tC.expected, result)
		})
	}
}

func TestGetPermissions(t *testing.T) {
	testCases := []struct {
		name                   string
		role                   string
		expected               map[string]map[string]any
		mockResourceAuthorizer func(m *MockResourceAuthorizer)
		resourceAuthorizers    ResourceAuthorizerMap
	}{
		{
			name: "Should return the correct permissions",
			role: ROLE_TOURNAMENT_DIRECTOR.ToString(),
			expected: map[string]map[string]any{
				"resource": {
					"create": true,
					"edit":   false,
					"get":    true,
					"list":   true,
					"subresource": map[string]any{
						"create": true,
						"delete": false,
						"get":    true,
						"list":   true,
					},
				},
			},
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("GetPermissions", mock.Anything).Return(map[string]any{
					"create": true,
					"edit":   false,
					"get":    true,
					"list":   true,
					"subresource": map[string]any{
						"create": true,
						"delete": false,
						"get":    true,
						"list":   true,
					},
				})
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			db, mock, err := database.NewMockDatabase()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}

			rows := sqlmock.NewRows([]string{"username", "password", "role"}).AddRow("testuser", "password", tC.role)
			mock.ExpectQuery("^SELECT \\* FROM \"logins\" WHERE \"logins\".\"username\" = \\$1$").WillReturnRows(rows)

			tC.mockResourceAuthorizer(tC.resourceAuthorizers["resource"].(*MockResourceAuthorizer))

			authService, err := NewAuthorizationService(db, "testuser", tC.resourceAuthorizers)
			if err != nil {
				t.Fatalf("NewAuthorizationService constructor should not have errored!")
			}

			result := authService.GetPermissions()
			assert.Equal(t, tC.expected, result)
		})
	}
}
