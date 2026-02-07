package middleware

import (
	"api/internal/authorization"
	"api/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UseAuthorization(action string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		role := ctx.GetString("role")
		if role == "" {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError("An error occurred during authorization."))
			return
		}

		authSvc := authorization.NewAuthorizationService(role, authorization.DefaultAuthorizerMap)

		authorized := authSvc.IsAuthorized(action)
		if !authorized {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Forbidden("You do not have permission to perform this action."))
			return
		}

		ctx.Next()
	}
}
