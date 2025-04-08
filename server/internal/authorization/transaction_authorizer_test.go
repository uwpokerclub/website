package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionAuthorizer(t *testing.T) {
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
			action: "get",
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
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_PRESIDENT.ToString(), expected: false},
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
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
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
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
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
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
				{role: ROLE_SECRETARY.ToString(), expected: false},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "edit",
		},
		{
			name: "Delete Authorized",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_BOT.ToString(), expected: false},
				{role: ROLE_EXECUTIVE.ToString(), expected: false},
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
				{role: ROLE_SECRETARY.ToString(), expected: false},
				{role: ROLE_TREASURER.ToString(), expected: true},
				{role: ROLE_VICE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_PRESIDENT.ToString(), expected: false},
				{role: ROLE_WEBMASTER.ToString(), expected: true},
			},
			action: "delete",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewTransactionAuthorizer()
			for _, role := range tC.roles {
				result := svc.IsAuthorized(role.role, tC.action)
				assert.Equal(t, role.expected, result)
			}
		})
	}
}

func TestTransactionAuthorizer_GetPermissions(t *testing.T) {
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
				"get":    false,
				"list":   false,
				"edit":   false,
				"delete": false,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewTransactionAuthorizer()
			permissions := svc.GetPermissions(tC.role)
			assert.Equal(t, tC.expected, permissions)
		})
	}
}
