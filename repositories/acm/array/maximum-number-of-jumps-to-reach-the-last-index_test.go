package array

import "testing"

func Test_maximumJumps(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums   []int
		target int
		want   int
	}{
		{
			nums:   []int{0, 2, 1, 3},
			target: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maximumJumps(tt.nums, tt.target)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maximumJumps() = %v, want %v", got, tt.want)
			}
		})
	}
}
