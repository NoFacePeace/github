package array

import "testing"

func Test_maxNumOfMarkedIndices(t *testing.T) {
	type args struct {
		nums []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// {
		// 	name: "test",
		// 	args: args{
		// 		nums: []int{42, 83, 48, 10, 24, 55, 9, 100, 10, 17, 17, 99, 51, 32, 16, 98, 99, 31, 28, 68, 71, 14, 64, 29, 15, 40},
		// 	},
		// 	want: 26,
		// },
		{
			name: "test",
			args: args{
				nums: []int{2, 5, 4, 9},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxNumOfMarkedIndices(tt.args.nums); got != tt.want {
				t.Errorf("maxNumOfMarkedIndices() = %v, want %v", got, tt.want)
			}
		})
	}
}
