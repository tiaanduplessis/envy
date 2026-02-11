package util

import "fmt"

// MaskValue masks a string value for display, showing only the first
// and last characters for strings longer than 3 characters.
func MaskValue(s string) string {
	if s == "" {
		return "(empty)"
	}
	if len(s) <= 3 {
		return "****"
	}
	return string(s[0]) + "****" + string(s[len(s)-1])
}

func FormatKeyValue(key, value string, reveal bool) string {
	if reveal {
		return fmt.Sprintf("%s=%s", key, value)
	}
	return fmt.Sprintf("%s=%s", key, MaskValue(value))
}
