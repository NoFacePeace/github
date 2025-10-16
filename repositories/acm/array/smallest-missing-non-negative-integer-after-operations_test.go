package array

import "testing"

func Test_findSmallestInteger(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums  []int
		value int
		want  int
	}{
		{
			nums: []int{1, -10, 7, 13, 6, 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findSmallestInteger(tt.nums, tt.value)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("findSmallestInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}
