package models

import "time"

const (
	FacultyAHS         = "AHS"
	FacultyArts        = "Arts"
	FacultyEngineering = "Engineering"
	FacultyEnvironment = "Environment"
	FacultyMath        = "Math"
	FacultyScience     = "Science"
)

type User struct {
	ID        uint64    `json:"id" binding:"required" gorm:"type:bigint;primaryKey;autoIncrement:false"`
	FirstName string    `json:"firstName" binding:"required"`
	LastName  string    `json:"lastName" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	Faculty   string    `json:"faculty" binding:"oneof=AHS Arts Engineering Environment Math Science"`
	QuestID   string    `json:"questId"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:timestamp;not null;default:LOCALTIMESTAMP"`
} //@name Member

type CreateUserRequest struct {
	ID        uint64 `json:"id" binding:"required"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Faculty   string `json:"faculty" binding:"oneof=AHS Arts Engineering Environment Math Science"`
	QuestID   string `json:"questId"`
} //@name CreateMemberRequest

type UpdateUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Faculty   string `json:"faculty" binding:"omitempty,oneof=AHS Arts Engineering Environment Math Science"`
	QuestID   string `json:"questId"`
} //@name UpdateMemberRequest

type ListUsersFilter struct {
	ID      *uint64
	Name    *string
	Email   *string
	Faculty *string
}
