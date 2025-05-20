package array

import "testing"

func Test_colorTheGrid(t *testing.T) {
	type args struct {
		m int
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				m: 1,
				n: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := colorTheGrid(tt.args.m, tt.args.n); got != tt.want {
				t.Errorf("colorTheGrid() = %v, want %v", got, tt.want)
			}
		})
	}
}
