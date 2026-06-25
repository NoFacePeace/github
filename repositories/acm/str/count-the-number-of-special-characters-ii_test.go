package str

import "testing"

func Test_numberOfSpecialChars(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		word string
		want int
	}{
		{
			word: "ehozyynxvvcpojdXUZVVRVPYSRQYZU",
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := numberOfSpecialChars(tt.word)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("numberOfSpecialChars() = %v, want %v", got, tt.want)
			}
		})
	}
}
