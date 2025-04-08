package models

import (
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID           uint64     `json:"id"`
	MembershipID uuid.UUID  `json:"membershipId" gorm:"type:uuid"`
	Membership   Membership `json:"membership"`
	EventID      uint64     `json:"eventId"`
	Placement    uint32     `json:"placement"`
	SignedOutAt  *time.Time `json:"signedOutAt"`
}

type CreateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      uint64    `json:"eventId" binding:"required"`
}

type UpdateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      uint64    `json:"eventId" binding:"required"`
	SignIn       bool
	SignOut      bool
}

type DeleteParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      uint64    `json:"eventId" binding:"required"`
}

type ListParticipantsResult struct {
	ID           uint64     `json:"id"`
	MembershipId uuid.UUID  `json:"membershipId"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	SignedOutAt  *time.Time `json:"signedOutAt"`
	Placement    uint32     `json:"placement"`
}
