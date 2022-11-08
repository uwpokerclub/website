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
