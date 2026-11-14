package array

import "testing"

func Test_threeSum(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want [][]int
	}{
		{
			nums: []int{1, 2, 0, 1, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := threeSum(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("threeSum() = %v, want %v", got, tt.want)
			}
		})
	}
}
