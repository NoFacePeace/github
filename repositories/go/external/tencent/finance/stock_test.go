package finance

import (
	"testing"
)

func TestListStocks(t *testing.T) {
	type args struct {
		options []ListStocksOption
	}
	tests := []struct {
		name    string
		args    args
		want    []Stock
		wantErr bool
	}{
		{
			name: "test",
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListStocks(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListStocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != listStocksDefaultCount {
				t.Errorf("ListStocks() got = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
