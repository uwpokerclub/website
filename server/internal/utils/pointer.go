package utils

// StringToPtr returns a pointer to value.
func StringToPtr(value string) *string {
	return &value
}

// Uint64ToPtr returns a pointer to value.
func Uint64ToPtr(value uint64) *uint64 {
	return &value
}

// BoolToPtr returns a pointer to value.
func BoolToPtr(value bool) *bool {
	return &value
}
