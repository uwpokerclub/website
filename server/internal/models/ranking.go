package models

import "github.com/google/uuid"

// SemesterRankingsView is the name of the database view that computes rankings
// with proper tie handling using RANK(). All ranking queries should use this
// constant rather than hardcoding the view name.
const SemesterRankingsView = "semester_rankings_view"

type Ranking struct {
	MembershipID uuid.UUID `json:"membershipId" gorm:"type:uuid;primaryKey"`
	Points       int32     `json:"points"`
	Attendance   int32     `json:"attendance" gorm:"not null;default:0"`
} //@name Ranking

func (Ranking) TableName() string {
	return "rankings"
}

type RankingResponse struct {
	ID        uint64 `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Points    int32  `json:"points"`
	Position  int32  `json:"position"`
} //@name RankingResponse

type GetRankingResponse struct {
	Points   int32 `json:"points"`
	Position int32 `json:"position"`
} //@name GetRankingResponse
