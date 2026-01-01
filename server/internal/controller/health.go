package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthController struct{}

func NewHealthController() Controller {
	return &healthController{}
}

func (h healthController) LoadRoutes(router *gin.RouterGroup) {
	router.GET("/health", h.healthCheck)
}

// healthCheck handles the health check endpoint.
// It returns a simple JSON response indicating the service is healthy.
//
// @Summary Health Check
// @Description Check the health of the API service
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h healthController) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
