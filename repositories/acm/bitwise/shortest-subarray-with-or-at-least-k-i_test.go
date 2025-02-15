package bitwise

import "testing"

func Test_minimumSubarrayLength(t *testing.T) {
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
				nums: []int{32, 2, 24, 1},
				k:    35,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minimumSubarrayLength(tt.args.nums, tt.args.k); got != tt.want {
				t.Errorf("minimumSubarrayLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
