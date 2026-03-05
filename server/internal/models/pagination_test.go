package models_test

import (
	"api/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestContext(queryString string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/?"+queryString, nil)
	return ctx
}

func TestParsePagination(t *testing.T) {
	t.Parallel()

	t.Run("no params sets defaults", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Limit)
		require.NotNil(t, p.Offset)
		require.Equal(t, models.MaxLimit, *p.Limit)
		require.Equal(t, 0, *p.Offset)
	})

	t.Run("only limit provided defaults offset to 0", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("limit=25")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Limit)
		require.NotNil(t, p.Offset)
		require.Equal(t, 25, *p.Limit)
		require.Equal(t, 0, *p.Offset)
	})

	t.Run("only offset provided defaults limit to MaxLimit", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("offset=10")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Limit)
		require.NotNil(t, p.Offset)
		require.Equal(t, models.MaxLimit, *p.Limit)
		require.Equal(t, 10, *p.Offset)
	})

	t.Run("both params provided uses provided values", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("limit=50&offset=20")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Limit)
		require.NotNil(t, p.Offset)
		require.Equal(t, 50, *p.Limit)
		require.Equal(t, 20, *p.Offset)
	})

	t.Run("limit exceeding max is capped", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("limit=500")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Limit)
		require.Equal(t, models.MaxLimit, *p.Limit)
	})

	t.Run("negative limit defaults to MaxLimit", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("limit=-1")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Limit)
		require.Equal(t, models.MaxLimit, *p.Limit)
	})

	t.Run("negative offset defaults to 0", func(t *testing.T) {
		t.Parallel()
		ctx := newTestContext("offset=-5")

		p, err := models.ParsePagination(ctx)
		require.NoError(t, err)
		require.NotNil(t, p.Offset)
		require.Equal(t, 0, *p.Offset)
	})
}
