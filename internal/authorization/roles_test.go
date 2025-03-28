package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRole(t *testing.T) {
	testCases := []struct {
		name     string
		role     role
		userRole string
		expected bool
	}{
		{
			name:     "Has the correct role",
			role:     ROLE_EXECUTIVE,
			userRole: "executive",
			expected: true,
		},
		{
			name:     "Has the incorrect role",
			role:     ROLE_EXECUTIVE,
			userRole: "unknown",
			expected: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			assert.Equal(t, tC.expected, HasRole(tC.role, tC.userRole))
		})
	}
}

func TestRoleToString(t *testing.T) {
	testCases := []struct {
		name     string
		role     role
		expected string
	}{
		{
			name:     "Executive",
			role:     ROLE_EXECUTIVE,
			expected: "executive",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			role := tC.role
			assert.Equal(t, tC.expected, role.ToString())
		})
	}
}

func TestStringToRole(t *testing.T) {
	testCases := []struct {
		desc     string
		role     string
		expected role
	}{
		{
			desc:     "Executive",
			role:     "executive",
			expected: ROLE_EXECUTIVE,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, stringToRole(tC.role))
		})
	}
}
