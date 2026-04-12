package dp

import "testing"

func Test_minimumDistance(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		word string
		want int
	}{
		{
			word: "CAKE",
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minimumDistance(tt.word)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minimumDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
