package array

import "testing"

func Test_trapRainWater(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		heightMap [][]int
		want      int
	}{
		{
			heightMap: [][]int{
				{78, 16, 94, 36},
				{87, 93, 50, 22},
				{63, 28, 91, 60},
				{64, 27, 41, 27},
				{73, 37, 12, 69},
				{68, 30, 83, 31},
				{63, 24, 68, 36}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trapRainWater(tt.heightMap)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("trapRainWater() = %v, want %v", got, tt.want)
			}
		})
	}
}
