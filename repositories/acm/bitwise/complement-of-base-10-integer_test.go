package bitwise

import "testing"

func Test_bitwiseComplement(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n    int
		want int
	}{
		{
			n:    5,
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bitwiseComplement(tt.n)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("bitwiseComplement() = %v, want %v", got, tt.want)
			}
		})
	}
}
