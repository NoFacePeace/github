package array

import "testing"

func Test_minimumHammingDistance(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		source       []int
		target       []int
		allowedSwaps [][]int
		want         int
	}{
		{
			source:       []int{1, 2, 3, 4},
			target:       []int{1, 3, 2, 4},
			allowedSwaps: [][]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minimumHammingDistance(tt.source, tt.target, tt.allowedSwaps)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minimumHammingDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
