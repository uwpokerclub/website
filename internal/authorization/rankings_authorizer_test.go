package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRankingsAuthorizer(t *testing.T) {
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
			action:   "get",
			expected: false,
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
			name:     "Export Authorized",
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "export",
			expected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			svc := NewRankingsAuthorizer()
			result := svc.IsAuthorized(tC.role, tC.action)
			assert.Equal(t, tC.expected, result)
		})
	}
}
