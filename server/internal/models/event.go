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
	ID               int32         `json:"id" gorm:"type:integer;primaryKey;autoIncrement"`
	Name             string        `json:"name"`
	Format           string        `json:"format"`
	Notes            string        `json:"notes"`
	SemesterID       uuid.UUID     `json:"semesterId" gorm:"type:uuid"`
	Semester         Semester      `json:"semester"`
	StartDate        time.Time     `json:"startDate" gorm:"not null;default:CURRENT_TIMESTAMP"`
	State            uint8         `json:"state" gorm:"default:0"`
	StructureID      int32         `json:"structureId" gorm:"type:integer;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Structure        Structure     `json:"structure"`
	Rebuys           uint8         `json:"rebuys" gorm:"not null;default:0"`
	PointsMultiplier float32       `json:"pointsMultiplier" gorm:"not null;default:1"`
	Entries          []Participant `json:"entries" gorm:"foreignKey:EventID"`
}

type CreateEventRequest struct {
	Name             string    `json:"name" binding:"required"`
	Format           string    `json:"format" binding:"required"`
	Notes            string    `json:"notes"`
	SemesterID       string    `json:"semesterId" binding:"required"`
	StartDate        time.Time `json:"startDate" binding:"required"`
	StructureID      int32     `json:"structureId" binding:"required"`
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
	ID         int32     `json:"id"`
	Name       string    `json:"name"`
	Format     string    `json:"format"`
	Notes      string    `json:"notes"`
	SemesterID string    `json:"semesterId"`
	StartDate  time.Time `json:"startDate"`
	State      uint8     `json:"state"`
	Count      int32     `json:"count"`
}
