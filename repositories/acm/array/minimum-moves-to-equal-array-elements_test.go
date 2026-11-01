package array

import "testing"

func Test_minMoves(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{
			name: "mixed signs",
			nums: []int{-100, 0, 100},
			want: 300,
		},
		{
			name: "ascending values",
			nums: []int{1, 2, 3},
			want: 3,
		},
		{
			name: "equal values",
			nums: []int{2, 2, 2},
			want: 0,
		},
		{
			name: "single value",
			nums: []int{5},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minMoves(tt.nums)
			if got != tt.want {
				t.Errorf("minMoves() = %v, want %v", got, tt.want)
			}
		})
	}
}
