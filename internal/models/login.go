package models

type Login struct {
	Username string `json:"username" binding:"required" gorm:"primaryKey"`
	Password string `json:"password" binding:"required"`
}
