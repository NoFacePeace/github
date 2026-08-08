package hash

import "testing"

func Test_validSequence(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		word1 string
		word2 string
		want  []int
	}{
		{
			word1: "bfcbbeeveebadfa",
			word2: "eeeee",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validSequence(tt.word1, tt.word2)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("validSequence() = %v, want %v", got, tt.want)
			}
		})
	}
}
