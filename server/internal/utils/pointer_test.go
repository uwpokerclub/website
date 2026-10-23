package utils

import "testing"

func TestBoolToPtr(t *testing.T) {
	for _, value := range []bool{false, true} {
		got := BoolToPtr(value)
		if got == nil || *got != value {
			t.Fatalf("BoolToPtr(%t) = %v, want pointer to %t", value, got, value)
		}
	}
}
