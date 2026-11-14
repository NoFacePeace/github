package array

import "testing"

func Test_subarraySum(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		k    int
		want int
	}{
		{
			nums: []int{1, 1, 1},
			k:    2,
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subarraySum(tt.nums, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("subarraySum() = %v, want %v", got, tt.want)
			}
		})
	}
}
