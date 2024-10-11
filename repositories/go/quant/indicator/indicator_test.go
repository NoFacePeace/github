package indicator

import (
	"fmt"
	"log"
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
			got, err := Price(tt.args.code)
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

func TestCrossMax(t *testing.T) {
	ps, err := AllPrice("sh600941")
	if err != nil {
		t.Error(err)
	}
	type args struct {
		ps    []Point
		short []Point
		long  []Point
	}
	tests := []struct {
		name string
		args args
		want []Point
	}{
		{
			name: "test",
			args: args{
				ps:    ps,
				short: SMA(ps, 5),
				long:  SMA(ps, 20),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CrossMax(tt.args.ps, tt.args.short, tt.args.long); len(got) == 0 {
				t.Errorf("CrossMax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrossLast(t *testing.T) {
	ps, err := AllPrice("sh600941")
	if err != nil {
		t.Error(err)
		return
	}
	type args struct {
		ps    []Point
		short []Point
		long  []Point
		last  int
	}
	tests := []struct {
		name string
		args args
		want []Point
	}{
		{
			name: "test",
			args: args{
				ps:    ps,
				short: SMA(ps, 5),
				long:  SMA(ps, 20),
				last:  1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GoldenCrossLast(tt.args.ps, tt.args.short, tt.args.long, tt.args.last); len(got) == 0 {
				t.Errorf("CrossLast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSMABestCrossLastPercent(t *testing.T) {
	ps, err := AllPrice("sh600941")
	if err != nil {
		log.Fatal(err)
	}
	short := SMA(ps, 5)
	long := SMA(ps, 20)
	cross := GoldenCrossLast(ps, short, long, 1)
	win := WinPercent(cross, 0)
	fmt.Println(win)
}
