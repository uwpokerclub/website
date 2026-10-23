package utils

// StringToPtr returns a pointer to value.
func StringToPtr(value string) *string {
	return &value
}

// Uint64ToPtr returns a pointer to value.
func Uint64ToPtr(value uint64) *uint64 {
	return &value
}

// StringPtrValue returns the pointed-to value or the string zero value.
func StringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// Uint64PtrValue returns the pointed-to value or the uint64 zero value.
func Uint64PtrValue(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}

// BoolToPtr returns a pointer to value.
func BoolToPtr(value bool) *bool {
	return &value
}
