package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructureAuthorizer(t *testing.T) {
	testCases := []struct {
		name  string
		roles []struct {
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
			action: "edit",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewStructureAuthorizer()
			for _, role := range tC.roles {
				result := svc.IsAuthorized(role.role, tC.action)
				assert.Equal(t, role.expected, result, "Expected %s to be %v for action %s", role.role, role.expected, tC.action)
			}
		})
	}
}

func TestStructureAuthorizer_GetPermissions(t *testing.T) {
	testCases := []struct {
		name     string
		role     string
		expected map[string]any
	}{
		{
			name: "Should return correct permission map",
			role: "tournament_director",
			expected: map[string]any{
				"create": true,
				"get":    true,
				"list":   true,
				"edit":   true,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewStructureAuthorizer()
			permissions := svc.GetPermissions(tC.role)
			assert.Equal(t, tC.expected, permissions)
		})
	}
}
