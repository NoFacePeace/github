package math

import "testing"

func Test_largestTriangleArea(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		points [][]int
		want   float64
	}{
		{
			points: [][]int{{35, -23}, {-12, -48}, {-34, -40}, {21, -25}, {-35, -44}, {24, 1}, {16, -9}, {41, 4}, {-36, -49}, {42, -49}, {-37, -20}, {-35, 11}, {-2, -36}, {18, 21}, {18, 8}, {-24, 14}, {-23, -11}, {-8, 44}, {-19, -3}, {0, -10}, {-21, -4}, {23, 18}, {20, 11}, {-42, 24}, {6, -19}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := largestTriangleArea(tt.points)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("largestTriangleArea() = %v, want %v", got, tt.want)
			}
		})
	}
}
