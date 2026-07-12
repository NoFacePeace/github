package sort

import "testing"

func Test_arrayRankTransform(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		arr  []int
		want []int
	}{
		{
			arr: []int{40, 10, 20, 30},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := arrayRankTransform(tt.arr)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("arrayRankTransform() = %v, want %v", got, tt.want)
			}
		})
	}
}
