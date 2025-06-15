package str

import "testing"

func Test_maxDifference(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				s: "tzt",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxDifference(tt.args.s); got != tt.want {
				t.Errorf("maxDifference() = %v, want %v", got, tt.want)
			}
		})
	}
}
