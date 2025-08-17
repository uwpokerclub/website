package testutils

import (
	"api/internal/server"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewTestAPIServer creates an API server configured for testing
// The caller is responsible for:
// - Setting up the database (using testcontainers or other method)
// - Creating http.ResponseRecorder and http.Request for testing
// - Managing test cleanup
func NewTestAPIServer(db *gorm.DB) *gin.Engine {
	// Set gin to test mode to disable debug output
	gin.SetMode(gin.TestMode)

	// Create the API server with the provided database
	apiServer := server.NewAPIServer(db)

	// Return the gin router for direct use in tests
	return apiServer.Router
}
