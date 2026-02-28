package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Participant struct {
	ID           int32       `json:"id" gorm:"type:integer;primaryKey;autoIncrement"`
	MembershipID *uuid.UUID  `json:"membershipId" gorm:"type:uuid;uniqueIndex:idx_membership_event"`
	Membership   *Membership `json:"membership,omitempty" gorm:"constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`
	EventID      int32       `json:"eventId" gorm:"type:integer;not null;uniqueIndex:idx_membership_event"`
	Placement    uint16      `json:"placement"`
	SignedOutAt  *time.Time  `json:"signedOutAt"`
} //@name Participant

func (Participant) TableName() string {
	return "participants"
}

func (Participant) Preload(tx *gorm.DB) *gorm.DB {
	return tx.
		Preload("Membership").
		Preload("Membership.User").
		Preload("Membership.Semester").
		Preload("Membership.Ranking")
}

type CreateParticipantRequest struct {
	MembershipID uuid.UUID `json:"membershipId" binding:"required"`
	EventID      int32     `json:"eventId" binding:"required"`
} //@name CreateParticipantRequest

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
} //@name ListParticipantsResult

type CreateEntryResult struct {
	MembershipID uuid.UUID    `json:"membershipId"`
	Status       string       `json:"status"` // "created" or "error"
	Participant  *Participant `json:"participant,omitempty"`
	Error        string       `json:"error,omitempty"`
} //@name CreateEntryResult
