package sort

import "testing"

func Test_countDays(t *testing.T) {
	type args struct {
		days     int
		meetings [][]int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				days:     48,
				meetings: [][]int{{26, 39}, {46, 47}, {9, 33}, {6, 33}, {28, 40}, {37, 39}, {14, 45}, {13, 40}, {14, 17}, {12, 39}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countDays(tt.args.days, tt.args.meetings); got != tt.want {
				t.Errorf("countDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
