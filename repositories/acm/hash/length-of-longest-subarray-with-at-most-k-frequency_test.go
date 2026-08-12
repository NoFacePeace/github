package hash

import "testing"

func Test_maxSubarrayLength(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		k    int
		want int
	}{
		{
			nums: []int{1, 3, 1, 1, 2},
			k:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxSubarrayLength(tt.nums, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxSubarrayLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
