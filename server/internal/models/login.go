package models

type Login struct {
	Username string    `json:"username" binding:"required" gorm:"primaryKey"`
	Password string    `json:"password" binding:"required" gorm:"not null"`
	Role     string    `json:"role" binding:"oneof=bot executive tournament_director secretary treasurer vice_president president" gorm:"size:20;not null;default:executive"`
	Sessions []Session `json:"-" gorm:"foreignKey:Username;references:Username;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
} //@name Login

type NewSessionRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
} //@name NewSessionRequest

// LinkedMemberInfo represents the member information linked to a login
type LinkedMemberInfo struct {
	ID        uint64 `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
} //@name LinkedMemberInfo

// LoginWithMember represents a login with its linked member information
type LoginWithMember struct {
	Username     string            `json:"username"`
	Role         string            `json:"role"`
	LinkedMember *LinkedMemberInfo `json:"linkedMember"`
} //@name LoginWithMember

// ChangePasswordRequest represents the request body for changing a password
type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=8"`
} //@name ChangePasswordRequest

// CreateLoginRequest represents the request body for creating a new login
type CreateLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=bot executive tournament_director secretary treasurer vice_president president webmaster"`
} //@name CreateLoginRequest
