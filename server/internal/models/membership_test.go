package models_test

import (
	"api/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMembership_EligibleForFreeTrial(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		paid     bool
		exec     bool
		eligible bool
	}{
		{name: "unpaid non-executive is eligible", paid: false, exec: false, eligible: true},
		{name: "paid non-executive is not eligible", paid: true, exec: false, eligible: false},
		{name: "unpaid executive is not eligible", paid: false, exec: true, eligible: false},
		{name: "paid executive is not eligible", paid: true, exec: true, eligible: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := models.Membership{Paid: tc.paid, Executive: tc.exec}
			require.Equal(t, tc.eligible, m.EligibleForFreeTrial())
		})
	}
}
