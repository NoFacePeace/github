package search

import "testing"

func Test_champagneTower(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		poured      int
		query_row   int
		query_glass int
		want        float64
	}{
		{
			poured:      25,
			query_row:   6,
			query_glass: 1,
			want:        1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := champagneTower(tt.poured, tt.query_row, tt.query_glass)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("champagneTower() = %v, want %v", got, tt.want)
			}
		})
	}
}
