package stack

import "testing"

func Test_survivedRobotsHealths(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		positions  []int
		healths    []int
		directions string
		want       []int
	}{
		{
			positions:  []int{4, 6},
			healths:    []int{601, 973},
			directions: "RL",
			want:       []int{972},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := survivedRobotsHealths(tt.positions, tt.healths, tt.directions)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("survivedRobotsHealths() = %v, want %v", got, tt.want)
			}
		})
	}
}
