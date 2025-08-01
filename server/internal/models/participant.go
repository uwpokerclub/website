package models

import (
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID           uint       `json:"id" gorm:"type:serial"`
	MembershipID uuid.UUID  `json:"membershipId" gorm:"type:uuid;primaryKey"`
	Membership   Membership `json:"membership"`
	EventID      uint       `json:"eventId" gorm:"primaryKey"`
	Placement    uint16     `json:"placement"`
	SignedOutAt  *time.Time `json:"signedOutAt"`
}

type CreateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      uint      `json:"eventId" binding:"required"`
}

type UpdateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      uint      `json:"eventId" binding:"required"`
	SignIn       bool
	SignOut      bool
}

type DeleteParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      uint      `json:"eventId" binding:"required"`
}

type ListParticipantsResult struct {
	ID           uint       `json:"id"`
	MembershipId uuid.UUID  `json:"membershipId"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	SignedOutAt  *time.Time `json:"signedOutAt"`
	Placement    uint16     `json:"placement"`
}
