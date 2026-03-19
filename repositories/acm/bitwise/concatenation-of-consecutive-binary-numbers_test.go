package bitwise

import "testing"

func Test_concatenatedBinary(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n    int
		want int
	}{
		{
			n:    42,
			want: 727837408,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := concatenatedBinary(tt.n)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("concatenatedBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}
