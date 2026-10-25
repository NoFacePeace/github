package str

import "testing"

func Test_countDaysTogether(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		arriveAlice string
		leaveAlice  string
		arriveBob   string
		leaveBob    string
		want        int
	}{
		{
			arriveAlice: "04-18",
			leaveAlice:  "04-26",
			arriveBob:   "04-02",
			leaveBob:    "04-24",
			want:        7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countDaysTogether(tt.arriveAlice, tt.leaveAlice, tt.arriveBob, tt.leaveBob)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("countDaysTogether() = %v, want %v", got, tt.want)
			}
		})
	}
}
