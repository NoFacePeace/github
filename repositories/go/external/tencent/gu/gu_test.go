package gu

import (
	"reflect"
	"testing"
)

func TestHKMinute(t *testing.T) {
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
			args: args{
				code: "hk00700",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HKMinute(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("HKMinute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HKMinute() = %v, want %v", got, tt.want)
			}
		})
	}
}
