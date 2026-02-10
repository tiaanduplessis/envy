package util

import "testing"

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "(empty)"},
		{"single char", "a", "****"},
		{"two chars", "ab", "****"},
		{"three chars", "abc", "****"},
		{"four chars", "abcd", "a****d"},
		{"long value", "super-secret-key-123", "s****3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskValue(tt.input)
			if got != tt.want {
				t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatKeyValue(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		value  string
		reveal bool
		want   string
	}{
		{"revealed", "DB_HOST", "localhost", true, "DB_HOST=localhost"},
		{"masked short", "KEY", "abc", false, "KEY=****"},
		{"masked long", "API_KEY", "sk-12345", false, "API_KEY=s****5"},
		{"masked empty", "EMPTY", "", false, "EMPTY=(empty)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatKeyValue(tt.key, tt.value, tt.reveal)
			if got != tt.want {
				t.Errorf("FormatKeyValue(%q, %q, %v) = %q, want %q",
					tt.key, tt.value, tt.reveal, got, tt.want)
			}
		})
	}
}
