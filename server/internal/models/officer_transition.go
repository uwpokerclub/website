package models

import (
	"github.com/google/uuid"
	"time"
)

const (
	OfficerTransitionPending   = "pending"
	OfficerTransitionCompleted = "completed"
	OfficerTransitionCancelled = "cancelled"
)

// OfficerTransition records a proposed team; staging never changes roles.
type OfficerTransition struct {
	ID                    uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InitiatedBy           string     `json:"initiatedBy" gorm:"not null"`
	PresidentUsername     string     `json:"presidentUsername" gorm:"not null"`
	VicePresidentUsername string     `json:"vicePresidentUsername" gorm:"not null"`
	SecretaryUsername     string     `json:"secretaryUsername" gorm:"not null"`
	TreasurerUsername     string     `json:"treasurerUsername" gorm:"not null"`
	Status                string     `json:"status" gorm:"size:20;not null;default:pending"`
	CreatedAt             time.Time  `json:"createdAt" gorm:"type:timestamp;not null;default:LOCALTIMESTAMP"`
	ResolvedAt            *time.Time `json:"resolvedAt" gorm:"type:timestamp"`
	// Nominees is populated on staging so the caller can confirm the resolved
	// people, not merely the supplied Quest IDs. It is not persisted.
	Nominees map[string]OfficerTransitionNominee `json:"nominees,omitempty" gorm:"-"`
}

type OfficerTransitionNominee struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
type CreateOfficerTransitionRequest struct {
	PresidentQuestID     string `json:"presidentQuestId" binding:"required"`
	VicePresidentQuestID string `json:"vicePresidentQuestId" binding:"required"`
	SecretaryQuestID     string `json:"secretaryQuestId" binding:"required"`
	TreasurerQuestID     string `json:"treasurerQuestId" binding:"required"`
}

type ReissueOfficerTransitionRequest struct {
	Role string `json:"role" binding:"required,oneof=president vice_president secretary treasurer"`
}
