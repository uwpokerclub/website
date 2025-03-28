package middleware

import (
	"api/internal/authorization"
	"api/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UseAuthorization(db *gorm.DB, action string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		username := ctx.GetString("username")
		if username == "" {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.InternalServerError("An error occurred during authorization."))
			return
		}

		authSvc, err := authorization.NewAuthorizationService(db, username, authorization.ResourceAuthorizerMap{
			"login": authorization.NewLoginAuthorizer(),
			"user":  authorization.NewUserAuthorizer(),
			"semester": authorization.NewSemesterAuthorizer(authorization.ResourceAuthorizerMap{
				"rankings":    authorization.NewRankingsAuthorizer(),
				"transaction": authorization.NewTransactionAuthorizer(),
			}),
			"membership": authorization.NewMembershipAuthorizer(),
			"structure":  authorization.NewStructureAuthorizer(),
			"event": authorization.NewEventAuthorizer(authorization.ResourceAuthorizerMap{
				"participant": authorization.NewParticipantAuthorizer(),
			}),
		})
		if err != nil {
			ctx.AbortWithStatusJSON(err.(errors.APIErrorResponse).Code, err)
			return
		}

		authorized := authSvc.IsAuthorized(action)
		if !authorized {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Forbidden("You do not have permission to perform this action."))
			return
		}

		ctx.Next()
	}
}
