package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cr "api/cron"
	"api/internal/database"
	"api/internal/server"

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

		// Initialize cron tasks
		c := cron.New()
		c.AddFunc("@daily", cr.SessionCleanup(db))
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

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: serv.Router,
		}

		// Start the HTTP server in a goroutine
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Failed to start HTTP server: %s", err.Error())
			}
		}()

		log.Printf("Server started on port %s", port)

		// Wait for SIGINT or SIGTERM
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		signal.Stop(quit)

		log.Println("Shutting down server...")

		// Stop the cron scheduler from scheduling new jobs
		cronCtx := c.Stop()

		// Gracefully shut down the HTTP server
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer httpCancel()

		if err := srv.Shutdown(httpCtx); err != nil {
			log.Printf("HTTP server forced to shutdown: %s", err.Error())
		} else {
			log.Println("HTTP server shut down")
		}

		// Wait for any running cron jobs to complete
		cronWaitCtx, cronWaitCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cronWaitCancel()

		select {
		case <-cronCtx.Done():
			log.Println("Cron scheduler stopped")
		case <-cronWaitCtx.Done():
			log.Println("Timed out waiting for cron jobs to finish")
		}

		// Close the database connection
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Failed to get underlying database connection: %s", err.Error())
		} else {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Failed to close database connection: %s", err.Error())
			} else {
				log.Println("Database connection closed")
			}
		}

		log.Println("Server exited")
	},
}
