package pointer

import "testing"

func Test_longestBalanced(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want int
	}{
		{
			nums: []int{9, 20, 5, 11, 20, 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestBalanced(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("longestBalanced() = %v, want %v", got, tt.want)
			}
		})
	}
}
