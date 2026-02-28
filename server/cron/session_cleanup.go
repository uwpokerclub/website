package cron

import (
	"api/internal/models"
	"log"
	"time"

	"gorm.io/gorm"
)

// SessionCleanup is a cron task that runs daily and will remove all expired sessions from the database. This is
// meant to preserve space in the database and prevent long standing sessions from taking up most of our database space.
func SessionCleanup(db *gorm.DB) func() {
	return func() {
		res := db.Where("expires_at < ?", time.Now().UTC()).Delete(&models.Session{})
		if err := res.Error; err != nil {
			log.Printf("Failed to delete expired sessions from the database: %v", err)
			return
		}
		log.Printf("Session cleanup complete: deleted %d expired session(s)", res.RowsAffected)
	}
}
