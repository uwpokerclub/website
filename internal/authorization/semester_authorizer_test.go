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
		role                   string
		action                 string
		expected               bool
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
			name:                "Unknown Sub-Resource",
			resourceAuthorizers: ResourceAuthorizerMap{},
			role:                ROLE_EXECUTIVE.ToString(),
			action:              "resource.create",
			expected:            false,
		},
		{
			name: "Sub-Resource Unauthorized",
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("IsAuthorized", mock.Anything, mock.Anything).Return(false)
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "resource.create",
			expected: false,
		},
		{
			name: "Sub-Resource Authorized",
			mockResourceAuthorizer: func(m *MockResourceAuthorizer) {
				m.On("IsAuthorized", mock.Anything, mock.Anything).Return(true)
			},
			resourceAuthorizers: ResourceAuthorizerMap{
				"resource": &MockResourceAuthorizer{},
			},
			role:     ROLE_EXECUTIVE.ToString(),
			action:   "resource.create",
			expected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if tC.mockResourceAuthorizer != nil {
				tC.mockResourceAuthorizer(tC.resourceAuthorizers["resource"].(*MockResourceAuthorizer))
			}

			svc := NewSemesterAuthorizer(tC.resourceAuthorizers)
			result := svc.IsAuthorized(tC.role, tC.action)
			assert.Equal(t, tC.expected, result)
		})
	}
}
