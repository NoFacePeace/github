package array

import "testing"

func Test_mostBooked(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n        int
		meetings [][]int
		want     int
	}{
		{
			n:        4,
			meetings: [][]int{{18, 19}, {3, 12}, {17, 19}, {2, 13}, {7, 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mostBooked(tt.n, tt.meetings)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("mostBooked() = %v, want %v", got, tt.want)
			}
		})
	}
}
