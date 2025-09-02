package search

import "testing"

func Test_lenOfVDiagonal(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		grid [][]int
		want int
	}{
		{
			grid: [][]int{{2, 2, 2, 2, 2}, {2, 0, 2, 2, 0}, {2, 0, 1, 1, 0}, {1, 0, 2, 2, 2}, {2, 0, 0, 2, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lenOfVDiagonal(tt.grid)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("lenOfVDiagonal() = %v, want %v", got, tt.want)
			}
		})
	}
}
