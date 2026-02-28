package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MaxBodySize(limit int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit)
		ctx.Next()
	}
}
