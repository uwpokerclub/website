package models_test

import (
	"api/internal/models"
	"encoding/json"
	"testing"
	"time"

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

func TestMembership_TrialTrackingTimestampsAreInternal(t *testing.T) {
	started := time.Date(2026, 9, 10, 18, 0, 0, 0, time.UTC)
	converted := started.AddDate(0, 0, 7)
	data, err := json.Marshal(models.Membership{TrialStartedAt: &started, ConvertedAt: &converted})
	require.NoError(t, err)
	require.NotContains(t, string(data), "trialStartedAt")
	require.NotContains(t, string(data), "convertedAt")
}
