package array

import "testing"

func Test_minScoreTriangulation(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		values []int
		want   int
	}{
		{
			values: []int{3, 1, 4, 5, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minScoreTriangulation(tt.values)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minScoreTriangulation() = %v, want %v", got, tt.want)
			}
		})
	}
}
