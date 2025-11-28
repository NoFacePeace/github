package str

import "testing"

func Test_findLexSmallestString(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		a    int
		b    int
		want string
	}{
		{
			s: "74",
			a: 5,
			b: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findLexSmallestString(tt.s, tt.a, tt.b)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("findLexSmallestString() = %v, want %v", got, tt.want)
			}
		})
	}
}
