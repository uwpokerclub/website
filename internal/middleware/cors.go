package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "https://uwpokerclub.com")
	} else {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	}
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	ctx.Writer.Header().Set("Access-Control-Allow-Headers", "User-Agent, Keep-Alive, Content-Type, Content-Length, Accept-Encoding, Cache-Control")
	ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PUT, PATCH")

	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(204)
		return
	}

	ctx.Next()
}
