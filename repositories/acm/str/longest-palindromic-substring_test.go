package str

import "testing"

func Test_longestPalindrome(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want string
	}{
		{
			s:    "babad",
			want: "bab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestPalindrome(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("longestPalindrome() = %v, want %v", got, tt.want)
			}
		})
	}
}
