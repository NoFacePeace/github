package array

import "testing"

func Test_minCost(t *testing.T) {
	type args struct {
		basket1 []int
		basket2 []int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			args: args{
				basket1: []int{84, 80, 43, 8, 80, 88, 43, 14, 100, 88},
				basket2: []int{32, 32, 42, 68, 68, 100, 42, 84, 14, 8},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minCost(tt.args.basket1, tt.args.basket2); got != tt.want {
				t.Errorf("minCost() = %v, want %v", got, tt.want)
			}
		})
	}
}
