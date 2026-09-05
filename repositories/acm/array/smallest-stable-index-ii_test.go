package array

import "testing"

func Test_firstStableIndex(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		k    int
		want int
	}{
		{
			nums: []int{2, 0, 2},
			k:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstStableIndex(tt.nums, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("firstStableIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
