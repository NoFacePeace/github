package array

import "testing"

func Test_minimumEffort(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tasks [][]int
		want  int
	}{
		{
			tasks: [][]int{{1, 2}, {2, 4}, {4, 8}},
			want:  8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minimumEffort(tt.tasks)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minimumEffort() = %v, want %v", got, tt.want)
			}
		})
	}
}
