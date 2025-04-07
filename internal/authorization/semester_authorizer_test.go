package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSemesterAuthorizer(t *testing.T) {
	testCases := []struct {
		name                   string
		mockResourceAuthorizer func(m *MockResourceAuthorizer)
		resourceAuthorizers    ResourceAuthorizerMap
		roles                  []struct {
			role     string
			expected bool
		}
		action string
	}{
		{
			name: "No action",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
			},
			action: "",
		},
		{
			name: "No role",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: "", expected: false},
			},
			action: "create",
		},
		{
			name: "Create Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
				{role: ROLE_SECRETARY.ToString(), expected: false},
				{role: ROLE_TREASURER.ToString(), expected: false},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "create",
		},
		{
			name: "Get Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: true},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: true},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "get",
		},
		{
			name: "List Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: true},
				{role: ROLE_EXECUTIVE.ToString(), expected: true},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: true},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "list",
		},
		{
			name:                "Unknown Sub-Resource",
			resourceAuthorizers: ResourceAuthorizerMap{},
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
			},
			action: "resource.create",
		},
		{
			name: "Sub-Resource Unauthorized",
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("IsAuthorized", mock.Anything, mock.Anything).Return(false)
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
			},
			action: "resource.create",
		},
		{
			name: "Sub-Resource Authorized",
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("IsAuthorized", mock.Anything, mock.Anything).Return(true)
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: true},
			},
			action: "resource.create",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if tC.mockResourceAuthorizer != nil {
				tC.mockResourceAuthorizer(tC.resourceAuthorizers["resource"].(*MockResourceAuthorizer))
			}

			svc := NewSemesterAuthorizer(tC.resourceAuthorizers)
			for _, r := range tC.roles {
				result := svc.IsAuthorized(r.role, tC.action)
				assert.Equal(t, r.expected, result, "Expected %s to be %v for action %s", r.role, r.expected, tC.action)
			}
		})
	}
}

func TestSemesterAuthorizer_GetPermissions(t *testing.T) {
	testCases := []struct {
		name                   string
		role                   string
		expected               map[string]any
		resourceAuthorizers    ResourceAuthorizerMap
		mockResourceAuthorizer func(m *MockResourceAuthorizer)
	}{
		{
			name: "Should return correct permission map",
			role: "tournament_director",
			expected: map[string]any{
				"create": false,
				"get":    true,
				"list":   true,
				"rankings": map[string]any{
					"create": false,
					"get":    true,
					"list":   true,
				},
				"transaction": map[string]any{
					"create": false,
					"get":    true,
					"list":   true,
				},
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"rankings":    &MockResourceAuthorizer{},
				"transaction": &MockResourceAuthorizer{},
			},
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("GetPermissions", mock.Anything).Return(map[string]any{
					"create": false,
					"get":    true,
					"list":   true,
				})
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			tC.mockResourceAuthorizer(tC.resourceAuthorizers["rankings"].(*MockResourceAuthorizer))
			tC.mockResourceAuthorizer(tC.resourceAuthorizers["transaction"].(*MockResourceAuthorizer))
			svc := NewSemesterAuthorizer(tC.resourceAuthorizers)
			permissions := svc.GetPermissions(tC.role)
			assert.Equal(t, tC.expected, permissions)
		})
	}
}
