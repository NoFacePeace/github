package hash

import "testing"

func Test_findAllPeople(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		n           int
		meetings    [][]int
		firstPerson int
		want        []int
	}{
		{
			n:           6,
			meetings:    [][]int{{1, 2, 5}, {2, 3, 8}, {1, 5, 10}},
			firstPerson: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAllPeople(tt.n, tt.meetings, tt.firstPerson)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("findAllPeople() = %v, want %v", got, tt.want)
			}
		})
	}
}
