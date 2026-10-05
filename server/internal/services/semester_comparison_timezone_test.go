package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMustLoadTorontoCalendar(t *testing.T) {
	t.Run("loads the required Toronto zone", func(t *testing.T) {
		location := mustLoadTorontoCalendar()
		require.NotNil(t, location)
		require.Equal(t, "America/Toronto", location.String())
	})

	t.Run("fails fast when the required zone cannot be loaded", func(t *testing.T) {
		require.Panics(t, func() {
			mustLoadTorontoCalendarWith(func(string) (*time.Location, error) {
				return nil, errors.New("zone database unavailable")
			})
		})
	})
}
