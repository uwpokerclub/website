package models

import "github.com/google/uuid"

type Ranking struct {
	MembershipID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Points       int32
}

type RankingResponse struct {
	ID        uint64 `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Points    int32  `json:"points"`
}
