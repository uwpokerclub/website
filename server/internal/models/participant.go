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

// ListParticipantsFilter is the set of parameters that will be used to filter the
// list entries query. EventID must be set by the caller; the zero value for the
// embedded Pagination is the same as not paginating the result. Search filters by
// participant/user first name, last name, full name, or student ID (case-insensitive),
// requiring a join against the memberships and users tables.
type ListParticipantsFilter struct {
	Pagination

	// EventID is the ID of the event to list entries for.
	EventID int32

	// Search filters entries by participant name or student ID (case-insensitive).
	Search string
}

type CreateEntryResult struct {
	MembershipID uuid.UUID    `json:"membershipId"`
	Status       string       `json:"status"` // "created" or "error"
	Participant  *Participant `json:"participant,omitempty"`
	Error        string       `json:"error,omitempty"`
} //@name CreateEntryResult
