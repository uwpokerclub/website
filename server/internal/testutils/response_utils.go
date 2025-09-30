package testutils

import (
	"encoding/json"
	"fmt"
)

// FindByID finds an item in a slice by matching its ID field
// T must be a type with an ID field that matches the provided id
func FindByID[T any, K comparable](slice []T, id K, getID func(T) K) (*T, error) {
	for _, item := range slice {
		if getID(item) == id {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("item with ID %v not found", id)
}

// ToMapInterface converts any struct to map[string]any via JSON marshaling
// This is useful for building expected responses in tests
func ToMapInterface(data any) (map[string]any, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}

// FindEventByIDAsMap finds an event by ID and returns it as map[string]any
func FindEventByIDAsMap(id int32) (map[string]any, error) {
	event, err := FindEventById(id)
	if err != nil {
		return nil, err
	}
	return ToMapInterface(event)
}

// FindStructureByIDAsMap finds a structure by ID and returns it as map[string]any
func FindStructureByIDAsMap(id int32) (map[string]any, error) {
	structure, err := FindStructureById(id)
	if err != nil {
		return nil, err
	}
	return ToMapInterface(structure)
}

// FindSemesterByIDAsMap finds a semester by ID and returns it as map[string]any
func FindSemesterByIDAsMap(id string) (map[string]any, error) {
	semester, err := FindSemesterByID(id)
	if err != nil {
		return nil, err
	}
	return ToMapInterface(semester)
}

