package array

import "testing"

func Test_avoidFlood(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		rains []int
		want  []int
	}{
		{
			rains: []int{1, 2, 0, 0, 2, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := avoidFlood(tt.rains)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("avoidFlood() = %v, want %v", got, tt.want)
			}
		})
	}
}
