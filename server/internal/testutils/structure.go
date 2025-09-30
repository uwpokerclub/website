package testutils

import (
	"api/internal/models"
	"fmt"

	"gorm.io/gorm"
)

var TEST_STRUCTURES = []models.Structure{
	{
		ID:   1,
		Name: "Standard Structure",
		Blinds: []models.Blind{
			{Index: 0, Small: 10, Big: 20, Ante: 0, Time: 15},
			{Index: 1, Small: 20, Big: 40, Ante: 0, Time: 15},
			{Index: 2, Small: 30, Big: 60, Ante: 0, Time: 15},
			{Index: 3, Small: 50, Big: 100, Ante: 0, Time: 15},
			{Index: 4, Small: 75, Big: 150, Ante: 0, Time: 15},
		},
	},
}

func SeedStructures(db *gorm.DB) error {
	for _, structure := range TEST_STRUCTURES {
		if err := db.Create(&structure).Error; err != nil {
			return err
		}
	}

	return nil
}

func FindStructureById(id int32) (*models.Structure, error) {
	for _, structure := range TEST_STRUCTURES {
		if structure.ID == id {
			return &structure, nil
		}
	}

	return nil, fmt.Errorf("structure not found")
}

// CreateTestStructure creates a test structure with the given name
func CreateTestStructure(db *gorm.DB, name string) (*models.Structure, error) {
	structure := models.Structure{
		Name: name,
		Blinds: []models.Blind{
			{Index: 0, Small: 10, Big: 20, Ante: 0, Time: 15},
			{Index: 1, Small: 20, Big: 40, Ante: 0, Time: 15},
			{Index: 3, Small: 30, Big: 60, Ante: 0, Time: 15},
			{Index: 3, Small: 50, Big: 100, Ante: 0, Time: 15},
			{Index: 4, Small: 75, Big: 150, Ante: 0, Time: 15},
		},
	}

	if err := db.Create(&structure).Error; err != nil {
		return nil, err
	}

	return &structure, nil
}
