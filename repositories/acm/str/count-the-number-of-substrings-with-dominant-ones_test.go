package str

import "testing"

func Test_numberOfSubstrings(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want int
	}{
		{
			s:    "101101",
			want: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := numberOfSubstrings(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("numberOfSubstrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
