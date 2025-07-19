package tencent

import (
	"testing"
)

func Test_getPlate(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		want    []StockPlate
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				code: "sh601698",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPlate(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPlate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Error("got empty")
			}
		})
	}
}

func Test_getBoardRankList(t *testing.T) {
	type args struct {
		code   string
		offset int
		count  int
	}
	tests := []struct {
		name    string
		args    args
		want    []Stock
		want1   int
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				code:   BoardCode1,
				offset: 0,
				count:  40,
			},
			wantErr: false,
			want1:   5363,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getBoardRankList(tt.args.code, tt.args.offset, tt.args.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBoardRankList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("getBoardRankList() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getBoardRankList() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getRank(t *testing.T) {
	type args struct {
		boardType string
		offset    int
		count     int
	}
	tests := []struct {
		name    string
		args    args
		want    []Plate
		want1   int
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				boardType: BoardType2,
				offset:    0,
				count:     40,
			},
			want1: 124,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getRank(tt.args.boardType, tt.args.offset, tt.args.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRank() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("getRank() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getRank() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getKline(t *testing.T) {
	type args struct {
		code   string
		limit  int
		ktype  string
		toDate string
	}
	tests := []struct {
		name    string
		args    args
		want    []Kline
		wantErr bool
	}{
		{
			name: "test plate",
			args: args{
				code:   "pt01801741",
				limit:  370,
				ktype:  "day",
				toDate: "",
			},
			wantErr: false,
		},
		{
			name: "test stock",
			args: args{
				code:   "sh601698",
				limit:  370,
				ktype:  "day",
				toDate: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKline(tt.args.code, tt.args.limit, tt.args.ktype, tt.args.toDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("got empty")
			}
		})
	}
}

func Test_getFullRank(t *testing.T) {
	type args struct {
		boardType string
	}
	tests := []struct {
		name    string
		args    args
		want    []Plate
		wantErr bool
	}{
		{
			args: args{
				boardType: BoardType2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFullRank(tt.args.boardType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFullRank() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 124 {
				t.Errorf("getFullRank() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFullBoardRankList(t *testing.T) {
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
			name: "test_getFullBoardRankList",
			args: args{
				code: BoardCode1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFullBoardRankList(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFullBoardRankList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 5363 {
				t.Errorf("got empty")
			}
		})
	}
}

func Test_getFullKline(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		want    []Kline
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				code: "sh601698",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFullKline(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFullKline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("getFullKline() = %v, want %v", got, tt.want)
			}
		})
	}
}
