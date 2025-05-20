package array

import "testing"

func Test_countSubarraysK(t *testing.T) {
	type args struct {
		nums []int
		k    int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// {
		// 	args: args{
		// 		nums: []int{1, 4, 2, 1},
		// 		k:    3,
		// 	},
		// },
		{
			args: args{
				nums: []int{1, 3, 2, 3, 3},
				k:    2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countSubarraysK(tt.args.nums, tt.args.k); got != tt.want {
				t.Errorf("countSubarraysK() = %v, want %v", got, tt.want)
			}
		})
	}
}
