package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	StartedAt time.Time `json:"startedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Username  string    `json:"username"`
}

type GetSessionResponse struct {
	Username    string                    `json:"username"`
	Role        string                    `json:"role"`
	Permissions map[string]map[string]any `json:"permissions"`
}
