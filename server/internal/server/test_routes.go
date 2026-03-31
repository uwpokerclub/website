//go:build e2e

package server

import (
	"api/internal/controller"

	"gorm.io/gorm"
)

func registerTestControllers(db *gorm.DB) []controller.Controller {
	return []controller.Controller{
		controller.NewTestResetController(db),
	}
}
