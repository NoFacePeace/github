package backtrack

import "testing"

func Test_makeTheIntegerZero(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		num1 int
		num2 int
		want int
	}{
		{
			num1: 3,
			num2: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeTheIntegerZero(tt.num1, tt.num2)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("makeTheIntegerZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
