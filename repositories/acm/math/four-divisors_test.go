package math

import "testing"

func Test_sumFourDivisors(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want int
	}{
		{
			nums: []int{21, 4, 7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sumFourDivisors(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("sumFourDivisors() = %v, want %v", got, tt.want)
			}
		})
	}
}
