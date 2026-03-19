package bitwise

import "testing"

func Test_findKthBit(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n    int
		k    int
		want byte
	}{
		{
			n:    4,
			k:    11,
			want: '1',
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findKthBit(tt.n, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("findKthBit() = %v, want %v", got, tt.want)
			}
		})
	}
}
