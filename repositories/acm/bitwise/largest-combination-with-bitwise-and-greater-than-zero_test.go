package bitwise

import "testing"

func Test_largestCombination(t *testing.T) {
	type args struct {
		candidates []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				candidates: []int{16, 17, 71, 62, 12, 24, 14},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := largestCombination(tt.args.candidates); got != tt.want {
				t.Errorf("largestCombination() = %v, want %v", got, tt.want)
			}
		})
	}
}
