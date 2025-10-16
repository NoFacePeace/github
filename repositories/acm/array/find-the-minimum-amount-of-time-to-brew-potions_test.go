package array

import "testing"

func Test_minTime(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		skill []int
		mana  []int
		want  int64
	}{
		{
			skill: []int{1, 5, 2, 4},
			mana:  []int{5, 1, 4, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minTime(tt.skill, tt.mana)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
