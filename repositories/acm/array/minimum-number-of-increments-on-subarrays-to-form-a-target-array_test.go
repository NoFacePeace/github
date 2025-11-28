package array

import "testing"

func Test_minNumberOperations(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		target []int
		want   int
	}{
		{
			target: []int{3, 1, 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minNumberOperations(tt.target)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minNumberOperations() = %v, want %v", got, tt.want)
			}
		})
	}
}
