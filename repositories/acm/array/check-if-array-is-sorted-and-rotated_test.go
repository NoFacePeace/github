package array

import "testing"

func Test_check(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want bool
	}{
		{
			nums: []int{3, 4, 5, 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("check() = %v, want %v", got, tt.want)
			}
		})
	}
}
