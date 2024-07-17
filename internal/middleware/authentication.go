package middleware

import (
	"net/http"
	"os"
	"strings"

	"api/internal/authentication"
	e "api/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UseAuthentication(db *gorm.DB) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var cookieKey string
		if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
			cookieKey = "uwpsc-session-id"
		} else {
			cookieKey = "uwpsc-dev-session-id"
		}
		cookie, err := ctx.Cookie(cookieKey)
		if err != nil {
			// Cookie not present in the request. Return 401
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
		}
		// Ensure cookie is a valid UUID
		err = uuid.Validate(cookie)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, e.Forbidden("Invalid session ID provided"))
			return
		}

		sessionID, _ := uuid.Parse(cookie)

		sessionManager := authentication.NewSessionManager(db)

		// Authenticate this session ID
		err = sessionManager.Authenticate(sessionID)
		if err != nil {
			ctx.AbortWithStatusJSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.Next()
	}
}
