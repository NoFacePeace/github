package array

import "testing"

func Test_maxIncreasingSubarrays(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want int
	}{
		{
			nums: []int{1, 2, 3, 4, 4, 4, 4, 5, 6, 7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxIncreasingSubarrays(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxIncreasingSubarrays() = %v, want %v", got, tt.want)
			}
		})
	}
}
