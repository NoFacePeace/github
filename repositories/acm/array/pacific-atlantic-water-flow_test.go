package array

import "testing"

func Test_pacificAtlantic(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		heights [][]int
		want    [][]int
	}{
		{
			heights: [][]int{{1, 2, 2, 3, 5}, {3, 2, 3, 4, 4}, {2, 4, 5, 3, 1}, {6, 7, 1, 4, 5}, {5, 1, 1, 2, 4}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pacificAtlantic(tt.heights)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("pacificAtlantic() = %v, want %v", got, tt.want)
			}
		})
	}
}
