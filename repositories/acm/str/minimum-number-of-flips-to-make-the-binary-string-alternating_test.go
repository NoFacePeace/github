package str

import "testing"

func Test_minFlips(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want int
	}{
		{
			s:    "111000",
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minFlips(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minFlips() = %v, want %v", got, tt.want)
			}
		})
	}
}
