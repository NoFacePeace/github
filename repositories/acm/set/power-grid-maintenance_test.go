package set

import "testing"

func Test_processQueries(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		c           int
		connections [][]int
		queries     [][]int
		want        []int
	}{
		{
			c:           5,
			connections: [][]int{{1, 2}, {2, 3}, {3, 4}, {4, 5}},
			queries:     [][]int{{1, 3}, {2, 1}, {1, 1}, {2, 2}, {1, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processQueries(tt.c, tt.connections, tt.queries)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("processQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
