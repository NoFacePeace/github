package hash

import "testing"

func Test_solveQueries(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums    []int
		queries []int
		want    []int
	}{
		{
			nums:    []int{1, 3, 1, 4, 1, 3, 2},
			queries: []int{0, 3, 5},
			want:    []int{2, -1, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := solveQueries(tt.nums, tt.queries)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("solveQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
