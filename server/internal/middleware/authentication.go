package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"api/internal/authentication"
	e "api/internal/errors"
	"api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UseAuthentication(st store.Store) func(ctx *gin.Context) {
	var cookieKey string
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		cookieKey = "uwpsc-session-id"
	} else {
		cookieKey = "uwpsc-dev-session-id"
	}

	sessionManager := authentication.NewSessionManager(st)

	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie(cookieKey)
		if err != nil {
			// Cookie not present in the request. Return 401
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
			return
		}
		// Ensure cookie is a valid UUID
		err = uuid.Validate(cookie)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, e.Forbidden("Invalid session ID provided"))
			return
		}

		sessionID, _ := uuid.Parse(cookie)

		// Authenticate this session ID
		session, err := sessionManager.Authenticate(sessionID)
		if err != nil {
			switch {
			case errors.Is(err, authentication.ErrSessionNotFound):
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
			case errors.Is(err, authentication.ErrSessionExpired):
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Session has expired. Please reauthenticate"))
			default:
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(err.Error()))
			}
			return
		}

		ctx.Set("username", session.Username)
		ctx.Set("role", session.Role)

		ctx.Next()
	}
}
