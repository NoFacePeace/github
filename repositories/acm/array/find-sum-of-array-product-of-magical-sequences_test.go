package array

import "testing"

func Test_magicalSum(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		m    int
		k    int
		nums []int
		want int
	}{
		{
			m:    5,
			k:    5,
			nums: []int{1, 10, 100, 10000, 1000000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := magicalSum(tt.m, tt.k, tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("magicalSum() = %v, want %v", got, tt.want)
			}
		})
	}
}
