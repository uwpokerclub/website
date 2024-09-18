package cmd

import (
	"api/internal/database"
	"api/internal/server"
	"fmt"
	"os"
	"strings"

	cr "api/cron"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var PORT string
var RUN_MIGRATIONS bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the UWPSC Admin API server",
	Run: func(cmd *cobra.Command, args []string) {
		// Establish connection to the database
		db, err := database.OpenConnection(RUN_MIGRATIONS)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open connection to the database: %s", err.Error())
			os.Exit(1)
		}

		// Turn on gorm debug mode to print SQL queries to the console in local development.
		if strings.ToLower(os.Getenv("ENVIRONMENT")) == "development" {
			db = db.Debug()
		}

		// Initialize cron tasks
		c := cron.New()
		c.AddFunc("@daily", cr.SessionCleanup(false))
		c.Start()

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

		serv.Run(port)
	},
}
