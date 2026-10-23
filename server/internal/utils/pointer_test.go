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

func TestStringToPtr(t *testing.T) {
	value := "roster"
	got := StringToPtr(value)
	if got == nil || *got != value {
		t.Fatalf("StringToPtr(%q) = %v, want pointer to %q", value, got, value)
	}
}

func TestUint64ToPtr(t *testing.T) {
	value := uint64(412)
	got := Uint64ToPtr(value)
	if got == nil || *got != value {
		t.Fatalf("Uint64ToPtr(%d) = %v, want pointer to %d", value, got, value)
	}
}
