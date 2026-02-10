package heap

import "testing"

func Test_minCost(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n     int
		edges [][]int
		want  int
	}{
		{
			n:     4,
			edges: [][]int{{0, 1, 3}, {3, 1, 1}, {2, 3, 4}, {0, 2, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minCost(tt.n, tt.edges)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minCost() = %v, want %v", got, tt.want)
			}
		})
	}
}
