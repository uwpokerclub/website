package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	e "api/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func UseAuthentication(ctx *gin.Context) {
	cookie, err := ctx.Cookie("pctoken")
	if err != nil {
		// Cookie not present in the request. Return 401
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
		return
	}

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
}
