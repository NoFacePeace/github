package tree

import "testing"

func Test_maximumBeauty(t *testing.T) {
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
				nums: []int{4, 6, 1, 2},
				k:    2,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maximumBeauty(tt.args.nums, tt.args.k); got != tt.want {
				t.Errorf("maximumBeauty() = %v, want %v", got, tt.want)
			}
		})
	}
}
