package dp

import "testing"

func Test_minimumDeleteSum(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s1   string
		s2   string
		want int
	}{
		{
			s1: "delete",
			s2: "leet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minimumDeleteSum(tt.s1, tt.s2)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minimumDeleteSum() = %v, want %v", got, tt.want)
			}
		})
	}
}
