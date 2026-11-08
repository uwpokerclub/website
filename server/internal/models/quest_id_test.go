package models

import "testing"

func TestNormalizeQuestID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "  AbC123  ", want: "abc123"},
		{input: " \t ", want: ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := NormalizeQuestID(test.input); got != test.want {
				t.Fatalf("NormalizeQuestID(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestValidateQuestID(t *testing.T) {
	tests := []struct {
		questID string
		valid   bool
	}{
		{questID: "abc123", valid: true},
		{questID: "", valid: false},
		{questID: "bad.id", valid: false},
		{questID: "bad-id", valid: false},
		{questID: "bad id", valid: false},
	}

	for _, test := range tests {
		t.Run(test.questID, func(t *testing.T) {
			if got := ValidateQuestID(test.questID); got != test.valid {
				t.Errorf("ValidateQuestID(%q) = %t, want %t", test.questID, got, test.valid)
			}
		})
	}
}
