//go:build e2e

package controller

import (
	_ "embed"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//go:embed testdata/seed.sql
var seedSQL string

type testResetController struct {
	db *gorm.DB
}

func NewTestResetController(db *gorm.DB) Controller {
	return &testResetController{db: db}
}

func (c *testResetController) LoadRoutes(router *gin.RouterGroup) {
	router.POST("/test/reset", c.resetDatabase)
}

func (c *testResetController) resetDatabase(ctx *gin.Context) {
	if strings.ToLower(os.Getenv("ENVIRONMENT")) != "test" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not available in this environment"})
		return
	}

	truncateSQL := `TRUNCATE blinds, events, memberships, participants,
		rankings, semesters, structures, transactions, users
		RESTART IDENTITY CASCADE`

	err := c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(truncateSQL).Error; err != nil {
			return err
		}
		return tx.Exec(seedSQL).Error
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
