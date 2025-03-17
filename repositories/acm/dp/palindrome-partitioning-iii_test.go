package dp

import "testing"

func Test_palindromePartition(t *testing.T) {
	type args struct {
		s string
		k int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				s: "aea",
				k: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := palindromePartition(tt.args.s, tt.args.k); got != tt.want {
				t.Errorf("palindromePartition() = %v, want %v", got, tt.want)
			}
		})
	}
}
