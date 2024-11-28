package array

import "testing"

func Test_minFlipsII(t *testing.T) {
	type args struct {
		grid [][]int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				// grid: [][]int{
				// 	{1, 0, 0},
				// 	{0, 1, 0},
				// 	{0, 0, 1},
				// },
				grid: [][]int{
					{0, 1},
					{0, 1},
					{0, 0},
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minFlipsII(tt.args.grid); got != tt.want {
				t.Errorf("minFlipsII() = %v, want %v", got, tt.want)
			}
		})
	}
}
