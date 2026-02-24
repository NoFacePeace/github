package bitwise

import "testing"

func Test_hasAllCodes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		k    int
		want bool
	}{
		{
			s:    "00110",
			k:    2,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAllCodes(tt.s, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("hasAllCodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
