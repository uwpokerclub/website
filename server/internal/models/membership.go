package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Membership struct {
	ID         uuid.UUID `json:"id"         gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uint64    `json:"userId"     gorm:"uniqueIndex:user_semester_unique"`
	User       *User     `json:"user"`
	SemesterID uuid.UUID `json:"semesterId" gorm:"type:uuid;uniqueIndex:user_semester_unique;index:idx_memberships_semester_id"`
	Semester   *Semester `json:"semester"`
	Paid       bool      `json:"paid"       gorm:"not null;default:false"`
	Discounted bool      `json:"discounted" gorm:"not null;default:false"`
	Ranking    *Ranking  `json:"ranking" gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
} //@name Membership

func (Membership) TableName() string {
	return "memberships"
}

func (Membership) Preload(tx *gorm.DB) *gorm.DB {
	return tx.Joins("User").Joins("Semester")
}

type CreateMembershipRequest struct {
	UserID     uint64 `json:"userId"     binding:"required"`
	SemesterID string `json:"semesterId" binding:"required"`
	Paid       bool   `json:"paid"       binding:"omitempty,required_with=Discounted"`
	Discounted bool   `json:"discounted" binding:"omitempty,required_with=Paid"`
}

type UpdateMembershipRequest struct {
	ID         uuid.UUID
	Paid       bool `json:"paid"       binding:"omitempty,required"`
	Discounted bool `json:"discounted" binding:"omitempty,required"`
}

type ListMembershipsResult struct {
	ID         uuid.UUID `json:"id"`
	UserID     uint64    `json:"userId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Paid       bool      `json:"paid"`
	Discounted bool      `json:"discounted"`
	Attendance int       `json:"attendance"`
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
}

type CreateMembershipRequestV2 struct {
	UserID     uint64 `json:"userId"     binding:"required"`
	Paid       bool   `json:"paid"       binding:"omitempty,required_with=Discounted"`
	Discounted bool   `json:"discounted" binding:"omitempty,required_with=Paid"`
} // @name CreateMembershipRequest

type UpdateMembershipRequestV2 struct {
	Paid       *bool `json:"paid"       binding:"omitempty"`
	Discounted *bool `json:"discounted" binding:"omitempty"`
} // @name UpdateMembershipRequest

// MembershipWithAttendance embeds Membership with computed attendance count
type MembershipWithAttendance struct {
	Membership
	Attendance int `json:"attendance"`
} // @name MembershipWithAttendance
