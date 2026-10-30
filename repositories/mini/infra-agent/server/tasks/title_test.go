package tasks

import "testing"

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		want     string
	}{
		{name: "empty", maxRunes: 5},
		{name: "below limit", input: "abc", maxRunes: 5, want: "abc"},
		{name: "at limit", input: "abcde", maxRunes: 5, want: "abcde"},
		{name: "above limit", input: "abcdef", maxRunes: 5, want: "ab..."},
		{name: "zero limit", input: "abcdef", want: "abcdef"},
		{name: "negative limit", input: "abcdef", maxRunes: -1, want: "abcdef"},
		{name: "one rune limit", input: "abcdef", maxRunes: 1, want: "a"},
		{name: "two rune limit", input: "abcdef", maxRunes: 2, want: "ab"},
		{name: "three rune limit", input: "abcdef", maxRunes: 3, want: "abc"},
		{name: "four rune limit", input: "abcdef", maxRunes: 4, want: "a..."},
		{name: "unicode at limit", input: "\u4f60\u597d", maxRunes: 2, want: "\u4f60\u597d"},
		{name: "unicode truncated", input: "\u4f60\u597d\u4e16\u754c\u4efb\u52a1", maxRunes: 5, want: "\u4f60\u597d..."},
		{name: "unicode small limit", input: "\u4f60\u597d\u4e16\u754c", maxRunes: 2, want: "\u4f60\u597d"},
		{name: "emoji", input: "\U0001f600abcdef", maxRunes: 5, want: "\U0001f600a..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateRunes(tt.input, tt.maxRunes); got != tt.want {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.input, tt.maxRunes, got, tt.want)
			}
		})
	}
}
