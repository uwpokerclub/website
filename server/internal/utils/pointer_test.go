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

func TestStringPtrValue(t *testing.T) {
	if got := StringPtrValue(nil); got != "" {
		t.Fatalf("StringPtrValue(nil) = %q, want empty string", got)
	}
	value := "roster"
	if got := StringPtrValue(&value); got != value {
		t.Fatalf("StringPtrValue(&value) = %q, want %q", got, value)
	}
}

func TestUint64PtrValue(t *testing.T) {
	if got := Uint64PtrValue(nil); got != 0 {
		t.Fatalf("Uint64PtrValue(nil) = %d, want 0", got)
	}
	value := uint64(412)
	if got := Uint64PtrValue(&value); got != value {
		t.Fatalf("Uint64PtrValue(&value) = %d, want %d", got, value)
	}
}
