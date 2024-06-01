package datetime

import (
	"testing"
	"time"
)

func TestIsHoliday(t *testing.T) {
	type args struct {
		t time.Time
	}
	holiday, _ := time.Parse(LayoutDate, "20240501")
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "no holiday",
			args: args{
				t: time.Now(),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "holiday",
			args: args{
				t: holiday,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsHoliday(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsHoliday() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsHoliday() = %v, want %v", got, tt.want)
			}
		})
	}
}
