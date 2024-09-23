package array

import "testing"

func Test_countWays(t *testing.T) {
	type args struct {
		nums []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				nums: []int{5, 0, 3, 4, 2, 1, 2, 4},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countWays(tt.args.nums); got != tt.want {
				t.Errorf("countWays() = %v, want %v", got, tt.want)
			}
		})
	}
}
