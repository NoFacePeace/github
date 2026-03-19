package array

import "testing"

func Test_minNumberOfSeconds(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		mountainHeight int
		workerTimes    []int
		want           int64
	}{
		{
			mountainHeight: 4,
			workerTimes:    []int{2, 1, 1},
			want:           3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minNumberOfSeconds(tt.mountainHeight, tt.workerTimes)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minNumberOfSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}
