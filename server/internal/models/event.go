package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventStateStarted = 0
	EventStateEnded   = 1
)

type Event struct {
	ID               uint64        `json:"id"`
	Name             string        `json:"name"`
	Format           string        `json:"format"`
	Notes            string        `json:"notes"`
	SemesterID       uuid.UUID     `json:"semesterId" gorm:"type:uuid"`
	Semester         Semester      `json:"semester"`
	StartDate        time.Time     `json:"startDate"`
	State            uint8         `json:"state"`
	StructureID      uint64        `json:"structureId"`
	Structure        Structure     `json:"structure"`
	Rebuys           uint8         `json:"rebuys"`
	PointsMultiplier float32       `json:"pointsMultiplier"`
	Entries          []Participant `json:"entries"`
}

type CreateEventRequest struct {
	Name             string    `json:"name" binding:"required"`
	Format           string    `json:"format" binding:"required"`
	Notes            string    `json:"notes"`
	SemesterID       string    `json:"semesterId" binding:"required"`
	StartDate        time.Time `json:"startDate" binding:"required"`
	StructureID      uint64    `json:"structureId" binding:"required"`
	PointsMultiplier float32   `json:"pointsMultiplier" binding:"required"`
}

type UpdateEventRequest struct {
	Name             *string    `json:"name"`
	Format           *string    `json:"format"`
	Notes            *string    `json:"notes"`
	StartDate        *time.Time `json:"startDate"`
	PointsMultiplier *float32   `json:"pointsMultiplier"`
}

type ListEventsResponse struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	Format     string    `json:"format"`
	Notes      string    `json:"notes"`
	SemesterID string    `json:"semesterId" gorm:"type:uuid"`
	StartDate  time.Time `json:"startDate"`
	State      uint8     `json:"state"`
	Count      int32     `json:"count"`
}
