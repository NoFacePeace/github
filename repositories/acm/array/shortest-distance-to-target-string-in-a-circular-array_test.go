package array

import "testing"

func Test_closestTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		words      []string
		target     string
		startIndex int
		want       int
	}{
		{
			words:      []string{"hello", "i", "am", "leetcode", "hello"},
			target:     "hello",
			startIndex: 1,
			want:       1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := closestTarget(tt.words, tt.target, tt.startIndex)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("closestTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}
