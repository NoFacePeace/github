package array

import "testing"

func Test_countPartitions(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want int
	}{
		{
			nums: []int{10, 10, 3, 7, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countPartitions(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("countPartitions() = %v, want %v", got, tt.want)
			}
		})
	}
}
