package bitwise

import "testing"

func Test_minimumSubarrayLengthII(t *testing.T) {
	type args struct {
		nums []int
		k    int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				nums: []int{1, 2},
				k:    0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minimumSubarrayLengthII(tt.args.nums, tt.args.k); got != tt.want {
				t.Errorf("minimumSubarrayLengthII() = %v, want %v", got, tt.want)
			}
		})
	}
}
