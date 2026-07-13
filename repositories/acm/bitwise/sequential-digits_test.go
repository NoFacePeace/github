package bitwise

import "testing"

func Test_sequentialDigits(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		low  int
		high int
		want []int
	}{
		{
			low:  100,
			high: 300,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sequentialDigits(tt.low, tt.high)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("sequentialDigits() = %v, want %v", got, tt.want)
			}
		})
	}
}
