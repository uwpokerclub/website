package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginAuthorizer(t *testing.T) {
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
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
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
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "create",
		},
		{
			name: "List Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "list",
		},
		{
			name: "Get Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "get",
		},
		{
			name: "Delete Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "delete",
		},
		{
			name: "Edit Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "edit",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewLoginAuthorizer()
			for _, r := range tC.roles {
				result := svc.IsAuthorized(r.role, tC.action)
				assert.Equal(t, r.expected, result, "Expected %s to be %v for action %s", r.role, r.expected, tC.action)
			}
		})
	}
}

func TestLoginAuthorizer_GetPermissions(t *testing.T) {
	testCases := []struct {
		name     string
		role     string
		expected map[string]any
	}{
		{
			name: "Should return correct permission map",
			role: "tournament_director",
			expected: map[string]any{
				"create": false,
				"list":   false,
				"get":    false,
				"delete": false,
				"edit":   false,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewLoginAuthorizer()
			permissions := svc.GetPermissions(tC.role)
			assert.Equal(t, tC.expected, permissions)
		})
	}
}
