package dp

import "testing"

func Test_minimumCostII(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		k    int
		dist int
		want int64
	}{
		{
			nums: []int{1, 6, 5, 7, 8, 7, 5},
			k:    5,
			dist: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minimumCostII(tt.nums, tt.k, tt.dist)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minimumCostII() = %v, want %v", got, tt.want)
			}
		})
	}
}
