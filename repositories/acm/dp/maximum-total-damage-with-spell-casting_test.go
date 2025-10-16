package dp

import "testing"

func Test_maximumTotalDamage(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		power []int
		want  int64
	}{
		{
			power: []int{1, 1, 3, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maximumTotalDamage(tt.power)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maximumTotalDamage() = %v, want %v", got, tt.want)
			}
		})
	}
}
