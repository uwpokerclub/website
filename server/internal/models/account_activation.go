package models

import (
	"time"

	"github.com/google/uuid"
)

// AccountActivation is a single-use, expiring password-activation token. TokenHash
// stores SHA-256(token), never the raw bearer secret.
type AccountActivation struct {
	TokenHash    []byte    `gorm:"primaryKey;type:bytea"`
	Username     string    `gorm:"not null;uniqueIndex:idx_account_activations_one_unused,where:used_at IS NULL"`
	ExpiresAt    time.Time `gorm:"not null"`
	UsedAt       *time.Time
	TransitionID *uuid.UUID         `json:"-" gorm:"type:uuid"`
	Transition   *OfficerTransition `json:"-" gorm:"foreignKey:TransitionID;references:ID;constraint:OnDelete:CASCADE"`
	Login        Login              `json:"-" gorm:"foreignKey:Username;references:Username;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

type VerifyActivationRequest struct {
	Token string `json:"token" binding:"required"`
}

type CompleteActivationRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type VerifyActivationResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
