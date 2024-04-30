package datetime

import (
	"testing"
	"time"
)

func TestIsWeekend(t *testing.T) {
	type args struct {
		t time.Time
	}
	Monday, _ := time.Parse(LayoutDateWithLine, "2024-04-29")
	Saturday, _ := time.Parse(LayoutDateWithLine, "2024-04-27")
	Sunday, _ := time.Parse(LayoutDateWithLine, "2024-04-28")
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Monday",
			args: args{
				t: Monday,
			},
			want: false,
		},
		{
			name: "Saturday",
			args: args{
				t: Saturday,
			},
			want: true,
		},
		{
			name: "Sunday",
			args: args{
				t: Sunday,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWeekend(tt.args.t); got != tt.want {
				t.Errorf("IsWeekend() = %v, want %v", got, tt.want)
			}
		})
	}
}
