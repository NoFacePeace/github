package hash

import "testing"

func Test_spellchecker(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		wordlist []string
		queries  []string
		want     []string
	}{
		{
			wordlist: []string{"KiTe", "kite", "hare", "Hare"},
			queries:  []string{"kite", "Kite", "KiTe", "Hare", "HARE", "Hear", "hear", "keti", "keet", "keto"},
			want:     []string{"kite", "KiTe", "KiTe", "Hare", "hare", "", "", "KiTe", "", "KiTe"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := spellchecker(tt.wordlist, tt.queries)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("spellchecker() = %v, want %v", got, tt.want)
			}
		})
	}
}
