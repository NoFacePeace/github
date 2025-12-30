package math

import "testing"

func Test_minSubarray(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		p    int
		want int
	}{
		{
			nums: []int{3, 1, 3},
			p:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minSubarray(tt.nums, tt.p)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minSubarray() = %v, want %v", got, tt.want)
			}
		})
	}
}
