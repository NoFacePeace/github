package array

import "testing"

func Test_longestSubarray(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want int
	}{
		{
			nums: []int{0, 1, 1, 1, 0, 1, 1, 0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestSubarray(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("longestSubarray() = %v, want %v", got, tt.want)
			}
		})
	}
}
