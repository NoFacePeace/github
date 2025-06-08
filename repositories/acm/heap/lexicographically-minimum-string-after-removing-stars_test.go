package heap

import "testing"

func Test_clearStars(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				s: "eydee**",
			},
			want: "eye",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clearStars(tt.args.s); got != tt.want {
				t.Errorf("clearStars() = %v, want %v", got, tt.want)
			}
		})
	}
}
