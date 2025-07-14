package finance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetKline(t *testing.T) {
	type args struct {
		code    string
		options []Option
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
				code:    "sh600941",
				options: []Option{WithAdjuct(NoneAdjust)},
			},
		},
		{
			name: "test 后复权",
			args: args{
				code:    "sh600941",
				options: []Option{WithAdjuct(AfterAdjust)},
			},
		},
		{
			name: "test 前复权",
			args: args{
				code:    "sh600941",
				options: []Option{WithAdjuct(BeforeAdjust)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetKline(tt.args.code, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetAllKline(t *testing.T) {
	type args struct {
		code    string
		options []Option
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
				// code: "sh600941",
				code: "pt01801055",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetAllKline(tt.args.code, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllKline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_getBoardRankList(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		want    []Stock
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				code: "aStock",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getBoardRankList(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBoardRankList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_getRank(t *testing.T) {
	data, err := getRank()
	require.Nil(t, err)
	require.Equal(t, len(data.RankList), 40)
}
