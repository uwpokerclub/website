package cron

import (
	"api/internal/database"
	"api/internal/models"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

// SessionCleanup is a cron task that runs daily and will remove all expired sessions from the database. This is
// meant to preserve space in the database and prevent long standing sessions from taking up most of our database space.
func SessionCleanup(testing bool) func() {
	return func() {
		// Open a new database connection, and do not run migrations. When this cron command is initialized the database
		// migrations will already have been ran, so we do not need to run them again.
		var db *gorm.DB
		var err error
		if !testing {
			db, err = database.OpenConnection(false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open connection to the database: %s", err.Error())
				os.Exit(1)
			}
		} else {
			db, err = database.OpenTestConnection()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open connection to the test database: %s", err.Error())
				os.Exit(1)
			}
		}

		// Close the database connection at the end of this function execution
		defer func() {
			sqlDB, err := db.DB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to retrieve the SQL DB from Gorm: %s", err.Error())
				os.Exit(1)
			}

			err = sqlDB.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to close the database: %s", err.Error())
				os.Exit(1)
			}
		}()

		// Get current time
		now := time.Now().UTC()

		// Delete all expired sessions
		res := db.Where("expires_at < ?", now.Format("2006-01-02 15:04:05")).Delete(&models.Session{})

		if err = res.Error; err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete expired sessions from the database: %s", err.Error())
		}
	}
}
