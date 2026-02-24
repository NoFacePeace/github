package bitwise

import "testing"

func Test_reverseBits(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n    int
		want int
	}{
		{
			name: "test",
			n:    43261596,
			want: 964176192,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reverseBits(tt.n)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("reverseBits() = %v, want %v", got, tt.want)
			}
		})
	}
}
