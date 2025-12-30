package dp

import "testing"

func Test_countPartitions(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		k    int
		want int
	}{
		{
			nums: []int{9, 4, 1, 3, 7},
			k:    4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countPartitions(tt.nums, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("countPartitions() = %v, want %v", got, tt.want)
			}
		})
	}
}
