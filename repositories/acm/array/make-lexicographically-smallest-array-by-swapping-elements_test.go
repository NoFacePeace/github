package array

import "testing"

func Test_lexicographicallySmallestArray(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums  []int
		limit int
		want  []int
	}{
		{
			nums:  []int{1, 81, 10, 79, 36, 2, 87, 12, 20, 77},
			limit: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexicographicallySmallestArray(tt.nums, tt.limit)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("lexicographicallySmallestArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
