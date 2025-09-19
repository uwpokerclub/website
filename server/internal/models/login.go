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
