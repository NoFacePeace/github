package array

import "testing"

func Test_predictTheWinner(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want bool
	}{
		{
			nums: []int{0, 0, 7, 6, 5, 6, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := predictTheWinner(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("predictTheWinner() = %v, want %v", got, tt.want)
			}
		})
	}
}
