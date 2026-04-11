package array

import "testing"

func Test_robotSim(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		commands  []int
		obstacles [][]int
		want      int
	}{
		{
			commands:  []int{7, -2, -2, 7, 5},
			obstacles: [][]int{},
			want:      49,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := robotSim(tt.commands, tt.obstacles)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("robotSim() = %v, want %v", got, tt.want)
			}
		})
	}
}
