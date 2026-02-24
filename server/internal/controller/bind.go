package controller

import (
	"errors"
	"net/http"

	apierrors "api/internal/errors"

	"github.com/gin-gonic/gin"
)

// BindJSON binds the request body as JSON into obj. On failure it writes the
// appropriate error response (413 for oversized bodies, 400 for other bind
// errors) and returns false.
func BindJSON(ctx *gin.Context, obj any) bool {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			ctx.AbortWithStatusJSON(http.StatusRequestEntityTooLarge,
				apierrors.RequestEntityTooLarge("request body too large"))
		} else {
			ctx.AbortWithStatusJSON(http.StatusBadRequest,
				apierrors.InvalidRequest(err.Error()))
		}
		return false
	}
	return true
}
