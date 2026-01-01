package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StartedAt time.Time `json:"startedAt" gorm:"not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null"`
	Username  string    `json:"username" gorm:"not null"`
} //@name Session

type GetSessionResponse struct {
	Username    string                    `json:"username"`
	Role        string                    `json:"role"`
	Permissions map[string]map[string]any `json:"permissions"`
} //@name GetSessionResponse
