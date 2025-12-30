package str

import "testing"

func Test_countCollisions(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		directions string
		want       int
	}{
		{
			directions: "RLRSLL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countCollisions(tt.directions)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("countCollisions() = %v, want %v", got, tt.want)
			}
		})
	}
}
