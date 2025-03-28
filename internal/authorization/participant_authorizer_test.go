package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParticipantAuthorizer(t *testing.T) {
	testCases := []struct {
		name     string
		role     string
		action   string
		expected bool
	}{
		{
			name:     "No action",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "",
			expected: false,
		},
		{
			name:     "No role",
			role:     "",
			action:   "",
			expected: false,
		},
		{
			name:     "Create Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "create",
			expected: true,
		},
		{
			name:     "Get Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "get",
			expected: true,
		},
		{
			name:     "List Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "list",
			expected: true,
		},
		{
			name:     "Sign-In Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "signin",
			expected: true,
		},
		{
			name:     "Sign-Out Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "signin",
			expected: true,
		},
		{
			name:     "Delete Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "delete",
			expected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewParticipantAuthorizer()
			result := svc.IsAuthorized(tC.role, tC.action)
			assert.Equal(t, tC.expected, result)
		})
	}
}
