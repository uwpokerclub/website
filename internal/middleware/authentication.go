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
		ctx.JSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
		return
	}

	token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		ctx.JSON(http.StatusForbidden, e.Forbidden(err.Error()))
		return
	}

	if token.Valid {
		ctx.Next()
	} else if errors.Is(err, jwt.ErrTokenMalformed) {
		ctx.JSON(http.StatusForbidden, e.Forbidden("Malformed token"))
	} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		ctx.JSON(http.StatusUnauthorized, e.Unauthorized("Authentication required"))
	} else {
		ctx.JSON(http.StatusInternalServerError, e.InternalServerError("Unknown error occurred authenticating request"))
	}
}
