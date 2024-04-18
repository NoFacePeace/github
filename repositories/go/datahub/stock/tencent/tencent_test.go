package tencent

import (
	"reflect"
	"testing"
)

func Test_getRank(t *testing.T) {
	type args struct {
		boardType string
	}
	tests := []struct {
		name    string
		args    args
		want    []Board
		wantErr bool
	}{
		{
			args: args{
				boardType: BoardType2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRank(tt.args.boardType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRank() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRank() = %v, want %v", got, tt.want)
			}
		})
	}
}
