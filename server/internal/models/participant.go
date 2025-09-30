package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Participant struct {
	ID           int32       `json:"id" gorm:"type:integer;unique;autoIncrement"`
	MembershipID uuid.UUID   `json:"membershipId" gorm:"type:uuid;primaryKey"`
	Membership   *Membership `json:"membership,omitempty"`
	EventID      int32       `json:"eventId" gorm:"type:integer;primaryKey;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Placement    uint16      `json:"placement"`
	SignedOutAt  *time.Time  `json:"signedOutAt"`
}

func (Participant) TableName() string {
	return "participants"
}

func (Participant) Preload(tx *gorm.DB) *gorm.DB {
	return tx.Joins("Membership", func(db *gorm.DB) *gorm.DB {
		return Membership{}.Preload(db)
	})
}

type CreateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      int32     `json:"eventId" binding:"required"`
}

type UpdateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      int32     `json:"eventId" binding:"required"`
	SignIn       bool
	SignOut      bool
}

type DeleteParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      int32     `json:"eventId" binding:"required"`
}

type ListParticipantsResult struct {
	ID           int32      `json:"id"`
	MembershipId uuid.UUID  `json:"membershipId"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	SignedOutAt  *time.Time `json:"signedOutAt"`
	Placement    uint16     `json:"placement"`
}
