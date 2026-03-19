package array

import "testing"

func Test_minSwaps(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		grid [][]int
		want int
	}{
		{
			grid: [][]int{
				{0, 0, 1},
				{1, 1, 0},
				{1, 0, 0},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minSwaps(tt.grid)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minSwaps() = %v, want %v", got, tt.want)
			}
		})
	}
}
