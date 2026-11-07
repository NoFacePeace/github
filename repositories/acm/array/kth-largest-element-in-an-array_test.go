package array

import "testing"

func Test_findKthLargest(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want int
	}{
		{
			name: "standard",
			nums: []int{3, 2, 1, 5, 6, 4},
			k:    2,
			want: 5,
		},
		{name: "largest", nums: []int{3, 2, 1}, k: 1, want: 3},
		{name: "unsorted_right_partition", nums: []int{1, 2, 3, 5, 4}, k: 2, want: 4},
		{name: "duplicates", nums: []int{3, 2, 3, 1, 2, 4, 5, 5, 6}, k: 4, want: 4},
		{name: "all_equal", nums: []int{2, 2, 2, 2}, k: 3, want: 2},
		{name: "single", nums: []int{7}, k: 1, want: 7},
		{name: "smallest", nums: []int{-1, -5, -3, 0}, k: 4, want: -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findKthLargest(tt.nums, tt.k)
			if got != tt.want {
				t.Errorf("findKthLargest() = %v, want %v", got, tt.want)
			}
		})
	}
}
