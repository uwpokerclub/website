package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventAuthorizer(t *testing.T) {
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
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: true},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
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
				{role: ROLE_BOT.ToString(), expected: false},
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
			name: "Edit Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: true},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "edit"},
		{
			name: "End Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "end",
		},
		{
			name: "Restart Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "restart",
		},
		{
			name: "Rebuy Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: true},
				{role: ROLE_SECRETARY.ToString(), expected: true},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_PRESIDENT.ToString(), expected: true},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			}, action: "rebuy",
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

			svc := NewEventAuthorizer(tC.resourceAuthorizers)
			for _, r := range tC.roles {
				result := svc.IsAuthorized(r.role, tC.action)
				assert.Equal(t, r.expected, result, "Expected %s to be %v for action %s", r.role, r.expected, tC.action)
			}
		})
	}
}

func TestEventAuthorizer_GetPermissions(t *testing.T) {
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
				"create":  true,
				"get":     true,
				"list":    true,
				"edit":    true,
				"end":     false,
				"restart": false,
				"rebuy":   true,
				"participant": map[string]any{
					"create": true,
					"get":    true,
					"list":   true,
					"delete": false,
				},
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"participant": &MockResourceAuthorizer{},
			},
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("GetPermissions", mock.Anything).Return(map[string]any{
					"create": true,
					"get":    true,
					"list":   true,
					"delete": false,
				})
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			tC.mockResourceAuthorizer(tC.resourceAuthorizers["participant"].(*MockResourceAuthorizer))
			svc := NewEventAuthorizer(tC.resourceAuthorizers)
			permissions := svc.GetPermissions(tC.role)
			assert.Equal(t, tC.expected, permissions)
		})
	}
}
