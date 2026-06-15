package store

import "api/internal/models"

// StructureRepistory is the interface for accessing the structures in the data store. It provides methods for creating, reading, updating, and deleting structures.
type StructureRepository interface {
	// Create creates a new structure in the data store.
	Create(structure *models.Structure) error

	// FindByID retrieves a structure from the data store by its ID.
	FindByID(id int32) (models.Structure, error)

	// List retrieves all structures from the data store.
	List(pagination *models.Pagination) ([]models.Structure, int64, error)

	// Update updates an existing structure in the data store.
	Update(structure *models.Structure) error

	// ReplaceBlindsByStructureID replaces all blinds for a structure with the given blinds.
	ReplaceBlindsByStructureID(structureID int32, blinds []models.Blind) error

	// Delete deletes a structure from the data store by its ID.
	Delete(id int32) error
}
