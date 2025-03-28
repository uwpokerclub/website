package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructureAuthorizer(t *testing.T) {
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
			action:   "create",
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
			name:     "Edit Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "edit",
			expected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewStructureAuthorizer()
			result := svc.IsAuthorized(tC.role, tC.action)
			assert.Equal(t, tC.expected, result)
		})
	}
}
