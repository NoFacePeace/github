package array

import "testing"

func Test_smallestRangeII(t *testing.T) {
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
			name: "test",
			args: args{
				nums: []int{2, 2, 7},
				k:    1,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := smallestRangeII(tt.args.nums, tt.args.k); got != tt.want {
				t.Errorf("smallestRangeII() = %v, want %v", got, tt.want)
			}
		})
	}
}
