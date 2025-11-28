package array

import "testing"

func Test_maxFrequency(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nums          []int
		k             int
		numOperations int
		want          int
	}{
		{
			nums:          []int{1, 90},
			k:             76,
			numOperations: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxFrequency(tt.nums, tt.k, tt.numOperations)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxFrequency() = %v, want %v", got, tt.want)
			}
		})
	}
}
