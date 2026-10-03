package dp

import "testing"

func Test_longestValidParentheses(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{
			name: "nested and adjacent pairs",
			s:    "()(())",
			want: 6,
		},
		{name: "empty", s: "", want: 0},
		{name: "unmatched opening", s: "(()", want: 2},
		{name: "unmatched closing", s: ")()())", want: 4},
		{name: "no valid pairs", s: "(((", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestValidParentheses(tt.s)
			if got != tt.want {
				t.Errorf("longestValidParentheses() = %v, want %v", got, tt.want)
			}
		})
	}
}
