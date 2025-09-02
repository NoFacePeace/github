package heap

import "testing"

func Test_maxAverageRatio(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		classes       [][]int
		extraStudents int
		want          float64
	}{
		{
			classes:       [][]int{{1, 2}, {3, 5}, {2, 2}},
			extraStudents: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxAverageRatio(tt.classes, tt.extraStudents)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxAverageRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}
