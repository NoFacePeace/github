package search

import "testing"

func Test_containsCycle(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		grid [][]byte
		want bool
	}{
		{
			name: "test case 1",
			grid: [][]byte{{'a', 'b', 'b'}, {'b', 'z', 'b'}, {'b', 'b', 'a'}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsCycle(tt.grid)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("containsCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}
