package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClockAuthorizer_IsAuthorized(t *testing.T) {
	testCases := []struct {
		name   string
		action string
		roles  []struct {
			role     string
			expected bool
		}
	}{
		{
			name:   "no action",
			action: "",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_TOURNAMENT_DIRECTOR.ToString(), expected: false},
			},
		},
		{
			name:   "no role",
			action: "get",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: "", expected: false},
			},
		},
		{
			name:   "get requires at least executive",
			action: "get",
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
		},
		{
			name:   "control requires at least tournament director",
			action: "control",
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
		},
		{
			name:   "unknown action",
			action: "delete",
			roles: []struct {
				role     string
				expected bool
			}{
				{role: ROLE_PRESIDENT.ToString(), expected: false},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewClockAuthorizer()
			for _, r := range tC.roles {
				result := svc.IsAuthorized(r.role, tC.action)
				assert.Equal(t, r.expected, result, "expected %s to be %v for action %s", r.role, r.expected, tC.action)
			}
		})
	}
}

func TestClockAuthorizer_GetPermissions(t *testing.T) {
	svc := NewClockAuthorizer()

	assert.Equal(t, map[string]any{
		"get":     true,
		"control": true,
	}, svc.GetPermissions(ROLE_TOURNAMENT_DIRECTOR.ToString()))

	assert.Equal(t, map[string]any{
		"get":     true,
		"control": false,
	}, svc.GetPermissions(ROLE_EXECUTIVE.ToString()))

	assert.Equal(t, map[string]any{
		"get":     false,
		"control": false,
	}, svc.GetPermissions(ROLE_BOT.ToString()))
}
