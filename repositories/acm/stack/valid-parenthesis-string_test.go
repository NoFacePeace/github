package stack

import "testing"

func Test_checkValidString(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want bool
	}{
		{
			s: "(((((*(()((((*((**(((()()*)()()()*((((**)())*)*)))))))(())(()))())((*()()(((()((()*(())*(()**)()(())",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkValidString(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("checkValidString() = %v, want %v", got, tt.want)
			}
		})
	}
}
