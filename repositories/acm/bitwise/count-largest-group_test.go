package bitwise

import "testing"

func Test_countLargestGroup(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				n: 13,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countLargestGroup(tt.args.n); got != tt.want {
				t.Errorf("countLargestGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
