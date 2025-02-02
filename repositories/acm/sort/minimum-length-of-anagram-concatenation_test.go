package sort

import "testing"

func Test_minAnagramLength(t *testing.T) {
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
				s: "abba",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minAnagramLength(tt.args.s); got != tt.want {
				t.Errorf("minAnagramLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
