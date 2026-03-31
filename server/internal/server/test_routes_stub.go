//go:build !e2e

package server

import (
	"api/internal/controller"

	"gorm.io/gorm"
)

func registerTestControllers(_ *gorm.DB) []controller.Controller {
	return nil
}
