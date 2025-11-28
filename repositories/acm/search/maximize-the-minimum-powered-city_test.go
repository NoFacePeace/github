package search

import "testing"

func Test_maxPower(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		stations []int
		r        int
		k        int
		want     int64
	}{
		{
			stations: []int{1, 2, 4, 5, 0},
			r:        1,
			k:        2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxPower(tt.stations, tt.r, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxPower() = %v, want %v", got, tt.want)
			}
		})
	}
}
