package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	EventStateStarted = 0
	EventStateEnded   = 1
)

type Event struct {
	ID               int32         `json:"id"                  gorm:"type:integer;primaryKey;autoIncrement"`
	Name             string        `json:"name"`
	Format           string        `json:"format"`
	Notes            string        `json:"notes"`
	SemesterID       uuid.UUID     `json:"semesterId"          gorm:"type:uuid"`
	Semester         *Semester     `json:"semester,omitempty"`
	StartDate        time.Time     `json:"startDate"           gorm:"not null;default:CURRENT_TIMESTAMP"`
	State            uint8         `json:"state"               gorm:"default:0"`
	StructureID      int32         `json:"structureId"         gorm:"type:integer;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Structure        *Structure    `json:"structure,omitempty"`
	Rebuys           uint8         `json:"rebuys"              gorm:"not null;default:0"`
	PointsMultiplier float32       `json:"pointsMultiplier"    gorm:"not null;default:1"`
	Entries          []Participant `json:"entries,omitempty"   gorm:"foreignKey:EventID"`
} //@name Event

func (Event) TableName() string {
	return "events"
}

type EventPreloadOptions struct {
	Semester  bool
	Structure bool
	Entries   bool
}

func (Event) Preload(tx *gorm.DB, options EventPreloadOptions) *gorm.DB {
	ret := tx

	if options.Semester {
		ret = ret.Joins("Semester")
	}

	if options.Structure {
		ret = ret.Preload("Structure", func(db *gorm.DB) *gorm.DB {
			return Structure{}.Preload(db, StructurePreloadOptions{Blinds: true})
		})
	}

	if options.Entries {
		ret = ret.Preload("Entries", func(db *gorm.DB) *gorm.DB {
			return Participant{}.Preload(db)
		})
	}

	return ret
}

func (event *Event) AfterSave(tx *gorm.DB) (err error) {
	err = event.Preload(tx, EventPreloadOptions{}).Find(event).Error
	return
}

type CreateEventRequest struct {
	Name             string    `json:"name"             binding:"required"`
	Format           string    `json:"format"           binding:"required"`
	Notes            string    `json:"notes"`
	SemesterID       string    `json:"semesterId"       binding:"required"`
	StartDate        time.Time `json:"startDate"        binding:"required"`
	StructureID      int32     `json:"structureId"      binding:"required"`
	PointsMultiplier float32   `json:"pointsMultiplier" binding:"required"`
} //@name CreateEventRequest

type UpdateEventRequest struct {
	Name             *string    `json:"name"`
	Format           *string    `json:"format"`
	Notes            *string    `json:"notes"`
	StartDate        *time.Time `json:"startDate"`
	PointsMultiplier *float32   `json:"pointsMultiplier"`
} //@name UpdateEventRequest

type UpdateEventRequestV2 struct {
	Name             *string    `json:"name,omitempty"             binding:"omitempty,min=1"                              example:"New Event Name"`
	Format           *string    `json:"format,omitempty"           binding:"omitempty,min=1"                              example:"No Limit Hold'em"`
	Notes            *string    `json:"notes,omitempty"            binding:"omitempty"                                    example:"Some notes about the event"`
	StartDate        *time.Time `json:"startDate,omitempty"        binding:"omitempty" example:"2023-10-01T18:00:00Z"`
	PointsMultiplier *float32   `json:"pointsMultiplier,omitempty" binding:"omitempty,gte=0"                              example:"1.5"`
} //@name UpdateEventRequestV2

type ListEventsResponse struct {
	ID         int32     `json:"id"`
	Name       string    `json:"name"`
	Format     string    `json:"format"`
	Notes      string    `json:"notes"`
	SemesterID string    `json:"semesterId"`
	StartDate  time.Time `json:"startDate"`
	State      uint8     `json:"state"`
	Count      int32     `json:"count"`
} //@name ListEventsResponse
