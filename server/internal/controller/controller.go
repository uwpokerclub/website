package controller

import "github.com/gin-gonic/gin"

// Controller defines the interface for a controller in the API server.
type Controller interface {
	LoadRoutes(router *gin.RouterGroup)
}
