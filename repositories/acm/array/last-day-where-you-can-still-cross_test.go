package array

import "testing"

func Test_latestDayToCross(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		row   int
		col   int
		cells [][]int
		want  int
	}{
		{
			row:   2,
			col:   2,
			cells: [][]int{{1, 1}, {2, 1}, {1, 2}, {2, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := latestDayToCross(tt.row, tt.col, tt.cells)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("latestDayToCross() = %v, want %v", got, tt.want)
			}
		})
	}
}
