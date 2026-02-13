package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAuthorizationService(t *testing.T) {
	svc := NewAuthorizationService("webmaster", nil)
	assert.NotNil(t, svc)
	assert.Equal(t, "webmaster", svc.role)
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
	testCases := []struct {
		name                   string
		mockResourceAuthorizer func(m *MockResourceAuthorizer)
		resourceAuthorizers    ResourceAuthorizerMap
		action                 string
		expected               bool
	}{
		{
			name:                "Empty action",
			resourceAuthorizers: ResourceAuthorizerMap{},
			action:              "",
			expected:            false,
		},
		{
			name:                "Only resource supplied",
			resourceAuthorizers: ResourceAuthorizerMap{},
			action:              "resource",
			expected:            false,
		},
		{
			name: "Unknown resource",
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			action:   "other.create",
			expected: false,
		},
		{
			name: "Resource Authorized",
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
			name: "Resource Unauthorized",
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
			if tC.mockResourceAuthorizer != nil {
				tC.mockResourceAuthorizer(tC.resourceAuthorizers["resource"].(*MockResourceAuthorizer))
			}

			authService := NewAuthorizationService("webmaster", tC.resourceAuthorizers)

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
			tC.mockResourceAuthorizer(tC.resourceAuthorizers["resource"].(*MockResourceAuthorizer))

			authService := NewAuthorizationService(tC.role, tC.resourceAuthorizers)

			result := authService.GetPermissions()
			assert.Equal(t, tC.expected, result)
		})
	}
}
