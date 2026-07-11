package search

import "testing"

func Test_countCompleteComponents(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n     int
		edges [][]int
		want  int
	}{
		{
			n:     5,
			edges: [][]int{{1, 0}, {2, 0}, {3, 0}, {4, 2}, {4, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countCompleteComponents(tt.n, tt.edges)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("countCompleteComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
