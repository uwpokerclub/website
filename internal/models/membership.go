package models

import "github.com/google/uuid"

type Membership struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID     uint64    `json:"userId"`
	SemesterID uuid.UUID `json:"semesterId" gorm:"type:uuid"`
	Paid       bool      `json:"paid"`
	Discounted bool      `json:"discounted"`
}

type CreateMembershipRequest struct {
	UserID     uint64 `json:"userId" binding:"required"`
	SemesterID string `json:"semesterId" binding:"required"`
	Paid       bool   `json:"paid" binding:"omitempty,required_with=Discounted"`
	Discounted bool   `json:"discounted" binding:"omitempty,required_with=Paid"`
}

type UpdateMembershipRequest struct {
	ID         uuid.UUID
	Paid       bool `json:"paid" binding:"omitempty,required"`
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
	// Limit is the the upper bound of results that will be returned by the query.
	// If this value is nil then no limit will be put on the query.
	Limit *int

	// Offset is the number of results to offset the query result by.
	// If this value is nil then no offset will be put on the query.
	Offset *int

	// SemesterID is the ID of the semester that you want to only list members from.
	// If this value is nil, then the query will return results from all semesters.
	SemesterID *uuid.UUID
}
