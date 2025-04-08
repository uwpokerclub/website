package models

type Login struct {
	Username string `json:"username" binding:"required" gorm:"primaryKey"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"oneof=bot executive tournament_director secretary treasurer vice_president president" gorm:"default:executive"`
}

type NewSessionRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
