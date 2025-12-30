package slidingwindow

import "testing"

func Test_maxProfit(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		prices   []int
		strategy []int
		k        int
		want     int64
	}{
		{
			prices:   []int{4, 7, 13},
			strategy: []int{-1, -1, 0},
			k:        2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxProfit(tt.prices, tt.strategy, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxProfit() = %v, want %v", got, tt.want)
			}
		})
	}
}
