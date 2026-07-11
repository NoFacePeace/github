package graph

import "testing"

func Test_findMaxPathScore(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		edges  [][]int
		online []bool
		k      int64
		want   int
	}{
		{
			edges:  [][]int{{0, 1, 5}, {1, 3, 10}, {0, 2, 3}, {2, 3, 4}},
			online: []bool{true, true, true, true},
			k:      10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMaxPathScore(tt.edges, tt.online, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("findMaxPathScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
