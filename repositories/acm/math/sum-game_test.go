package math

import "testing"

func Test_sumGame(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		num  string
		want bool
	}{
		{
			num: "9?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sumGame(tt.num)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("sumGame() = %v, want %v", got, tt.want)
			}
		})
	}
}
