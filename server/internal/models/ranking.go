package models

import "github.com/google/uuid"

type Ranking struct {
	MembershipID uuid.UUID `json:"membershipId" gorm:"type:uuid;primaryKey"`
	Points       int32     `json:"points"`
	Attendance   int32     `json:"attendance"`
}

type RankingResponse struct {
	ID        uint64 `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Points    int32  `json:"points"`
}

type GetRankingResponse struct {
	Points   int32 `json:"points"`
	Position int32 `json:"position"`
}
