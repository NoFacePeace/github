package indicator

import (
	"testing"
)

func TestPrice(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		want    []Point
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				code: "sh600941",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllPrice(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("Price() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("Price() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSMA(t *testing.T) {
	ps, err := AllPrice("sh600941")
	if err != nil {
		t.Error(err)
	}
	type args struct {
		ps     []Point
		window int
	}
	tests := []struct {
		name string
		args args
		want []Point
	}{
		{
			name: "test",
			args: args{
				ps:     ps,
				window: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SMA(tt.args.ps, tt.args.window)
			if len(got) == 0 {
				t.Errorf("SMA() = %v, want %v", got, tt.want)
			}
		})
	}
}