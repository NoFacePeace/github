package str

import "testing"

func Test_minimumSubstringsInPartition(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				s: "bccbaacabc",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minimumSubstringsInPartition(tt.args.s); got != tt.want {
				t.Errorf("minimumSubstringsInPartition() = %v, want %v", got, tt.want)
			}
		})
	}
}
