package hash

import "testing"

func Test_longestBalancedII(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{
			s: "abbac",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longestBalancedII(tt.s); got != tt.want {
				t.Errorf("longestBalancedSubstring(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
