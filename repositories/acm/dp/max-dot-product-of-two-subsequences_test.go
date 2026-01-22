package dp

import "testing"

func Test_maxDotProduct(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums1 []int
		nums2 []int
		want  int
	}{
		{
			nums1: []int{2, -4, -7, -9, -4},
			nums2: []int{-9, -1, 1, -2, -2, -4, 5, 10, 9, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxDotProduct(tt.nums1, tt.nums2)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxDotProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}
