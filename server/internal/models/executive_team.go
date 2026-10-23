package models

// ExecutiveTeamLogin is the public-safe account information attached to a roster member.
// It deliberately excludes credentials, sessions, activations, and transition data.
type ExecutiveTeamLogin struct {
	Username      string `json:"username"`
	Role          string `json:"role"`
	AccountStatus string `json:"accountStatus"`
} //@name ExecutiveTeamLogin

// ExecutiveTeamMember is one deduplicated person in a semester's roster.
type ExecutiveTeamMember struct {
	MemberID               uint64  `json:"memberId"`
	FirstName              string  `json:"firstName"`
	LastName               string  `json:"lastName"`
	QuestID                string  `json:"questId"`
	Username               *string `json:"username"`
	Role                   *string `json:"role"`
	AccountStatus          *string `json:"accountStatus"`
	ActivationOutstanding  bool    `json:"activationOutstanding"`
	HasLogin               bool    `json:"hasLogin"`
	HasExecutiveMembership bool    `json:"hasExecutiveMembership"`
} //@name ExecutiveTeamMember

// ExecutiveTeamResponse is the roster envelope. Members is always initialized
// to an empty slice so an empty roster serializes as [] rather than null.
type ExecutiveTeamResponse struct {
	Members []ExecutiveTeamMember `json:"members"`
} //@name ExecutiveTeamResponse
