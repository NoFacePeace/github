package dp

import "testing"

func Test_stoneGameII(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		piles []int
		want  int
	}{
		{
			piles: []int{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stoneGameII(tt.piles)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("stoneGameII() = %v, want %v", got, tt.want)
			}
		})
	}
}
