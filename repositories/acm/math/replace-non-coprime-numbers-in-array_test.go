package math

import "testing"

func Test_replaceNonCoprimes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums []int
		want []int
	}{
		{
			nums: []int{8303, 361, 8303, 361, 437, 361, 8303, 8303, 8303, 6859, 19, 19, 361, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 70121, 1271, 31, 961, 31, 7, 2009, 7, 2009, 2009, 49, 7, 7, 8897, 1519, 31, 1519, 217},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceNonCoprimes(tt.nums)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("replaceNonCoprimes() = %v, want %v", got, tt.want)
			}
		})
	}
}
