package array

import "testing"

func Test_findDiagonalOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		mat  [][]int
		want []int
	}{
		{
			mat: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDiagonalOrder(tt.mat)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("findDiagonalOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
