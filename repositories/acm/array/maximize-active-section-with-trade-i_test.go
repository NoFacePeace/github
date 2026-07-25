package array

import "testing"

func Test_maxActiveSectionsAfterTrade(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want int
	}{
		{
			s: "0100",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxActiveSectionsAfterTrade(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxActiveSectionsAfterTrade() = %v, want %v", got, tt.want)
			}
		})
	}
}
