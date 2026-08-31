package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Membership struct {
	ID                 uuid.UUID `json:"id"         gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             uint64    `json:"userId"     gorm:"uniqueIndex:user_semester_unique"`
	User               *User     `json:"user" gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	SemesterID         uuid.UUID `json:"semesterId" gorm:"type:uuid;uniqueIndex:user_semester_unique;index:idx_memberships_semester_id"`
	Semester           *Semester `json:"semester"`
	Paid               bool      `json:"paid"       gorm:"not null;default:false"`
	Discounted         bool      `json:"discounted" gorm:"not null;default:false"`
	Executive          bool      `json:"executive"  gorm:"not null;default:false"`
	// FreeTrialAvailable is only kept in sync for memberships eligible for the free trial in the
	// first place (see EligibleForFreeTrial). For an executive, this column keeps whatever value
	// it last held and must not be read on its own; always gate reads behind EligibleForFreeTrial.
	FreeTrialAvailable bool     `json:"freeTrialAvailable" gorm:"not null;default:true"`
	Ranking            *Ranking `json:"ranking" gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
} //@name Membership

// EligibleForFreeTrial reports whether the free-trial limit applies to this
// membership at all. Paid members have bought unlimited entry; executives are
// comped and were never on a trial to begin with.
func (m Membership) EligibleForFreeTrial() bool {
	return !m.Paid && !m.Executive
}

func (Membership) TableName() string {
	return "memberships"
}

func (Membership) Preload(tx *gorm.DB) *gorm.DB {
	return tx.Joins("User").Joins("Semester")
}

// ListMembershipsFilter is the set of parameters that will be used to filter the
// list memberships query. The zero value for ListMembershipsFilter is the same as
// not filtering the result.
type ListMembershipsFilter struct {
	Pagination

	// SemesterID is the ID of the semester that you want to only list members from.
	// If this value is nil, then the query will return results from all semesters.
	SemesterID *uuid.UUID

	// UserID will filter for memberships that are only held by this specified user.
	UserID *uint64

	// Search filters memberships by user first name, last name, email, or full name (case-insensitive).
	Search string

	// Name filters memberships by user first name, last name, or full name (case-insensitive partial match).
	Name *string

	// Email filters memberships by user email (case-insensitive partial match).
	Email *string

	// Faculty filters memberships by user faculty (exact match).
	Faculty *string

	// StudentID filters memberships by user ID (exact match as string).
	StudentID *string

	// Paid filters memberships by paid status.
	Paid *bool

	// Discounted filters memberships by discounted status.
	Discounted *bool
}

type CreateMembershipRequest struct {
	UserID     uint64 `json:"userId"     binding:"required"`
	Paid       bool   `json:"paid"       binding:"omitempty,required_with=Discounted"`
	Discounted bool   `json:"discounted" binding:"omitempty,required_with=Paid"`
} // @name CreateMembershipRequest

type UpdateMembershipRequest struct {
	Paid       *bool `json:"paid"       binding:"omitempty"`
	Discounted *bool `json:"discounted" binding:"omitempty"`
} // @name UpdateMembershipRequest

// MembershipWithAttendance embeds Membership with computed attendance count
type MembershipWithAttendance struct {
	Membership
	Attendance int `json:"attendance"`
} // @name MembershipWithAttendance
