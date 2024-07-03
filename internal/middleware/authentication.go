package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"api/internal/authentication"
	e "api/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UseAuthentication(db *gorm.DB) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// Check if cookie is a JWT token
		cookie, err := ctx.Cookie("pctoken")
		if err == nil {
			token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}

				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if token.Valid {
				ctx.Next()
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, e.Forbidden("Malformed token"))
			} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
			} else {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, e.InternalServerError(fmt.Sprintf("Unknown error occurred authenticating request: %s", err.Error())))
			}
			return
		}

		var cookieKey string
		if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
			cookieKey = "uwpsc-session-id"
		} else {
			cookieKey = "uwpsc-dev-session-id"
		}
		cookie, err = ctx.Cookie(cookieKey)
		if err == nil {
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

			return
		}

		// Cookie not present in the request. Return 401
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
	}
}
