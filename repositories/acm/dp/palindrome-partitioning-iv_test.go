package dp

import "testing"

func Test_checkPartitioning(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{
				s: "abcbdd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPartitioning(tt.args.s); got != tt.want {
				t.Errorf("checkPartitioning() = %v, want %v", got, tt.want)
			}
		})
	}
}
