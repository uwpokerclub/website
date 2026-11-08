package models

import (
	"regexp"
	"strings"
)

var questIDPattern = regexp.MustCompile(`^[a-z0-9]+$`)

// NormalizeQuestID prepares newly supplied Quest ID input for validation/storage.
func NormalizeQuestID(questID string) string {
	return strings.ToLower(strings.TrimSpace(questID))
}

// ValidateQuestID reports whether a normalized non-empty Quest ID is supported.
func ValidateQuestID(questID string) bool {
	return questIDPattern.MatchString(questID)
}
