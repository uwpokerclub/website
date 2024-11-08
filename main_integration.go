//go:build test
// +build test

package main

import (
	"api/internal/database"
	"api/internal/server"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var PORT = "5000"

func main() {
	db, err := database.OpenConnection(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open connection to the database: %s", err.Error())
		os.Exit(1)
	}

	// Turn on gorm debug mode to print SQL queries to the console in local development.
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "development" {
		db = db.Debug()
	}

	// Initialize the server
	serv := server.NewAPIServer(db)

	// Determine the port to run the server on. If only the PORT environment
	// variable is set, use that as the port. If a port is provided via a
	// a command flag, use this value instead
	port := os.Getenv("PORT")
	if port == "" {
		// Use command flag value instead
		port = PORT
	}

	serv.Router.POST("/database/reset", func(ctx *gin.Context) {
		resetSession := db.Session(&gorm.Session{AllowGlobalUpdate: true})

		oldEnvironment := os.Getenv("ENVIRONMENT")
		os.Setenv("ENVIRONMENT", "test")
		defer os.Setenv("ENVIRONMENT", oldEnvironment)

		err := database.WipeDB(resetSession)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Status(http.StatusOK)
	})

	serv.Run(port)
}
