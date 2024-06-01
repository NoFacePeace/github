package datetime

import (
	"reflect"
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

func TestYesterday(t *testing.T) {
	type args struct {
		ts []time.Time
	}
	now := time.Now()
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "test",
			args: args{
				[]time.Time{
					now,
				},
			},
			want: now.AddDate(0, 0, -1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Yesterday(tt.args.ts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Yesterday() = %v, want %v", got, tt.want)
			}
		})
	}
}
