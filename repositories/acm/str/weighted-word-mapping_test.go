package str

import "testing"

func Test_mapWordWeights(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		words   []string
		weights []int
		want    string
	}{
		{
			words:   []string{"abcd", "def", "xyz"},
			weights: []int{5, 3, 12, 14, 1, 2, 3, 2, 10, 6, 6, 9, 7, 8, 7, 10, 8, 9, 6, 9, 9, 8, 3, 7, 7, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapWordWeights(tt.words, tt.weights)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("mapWordWeights() = %v, want %v", got, tt.want)
			}
		})
	}
}
