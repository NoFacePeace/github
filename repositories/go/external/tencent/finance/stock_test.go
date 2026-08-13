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
			args: args{
				options: []ListStocksOption{
					WithListStocksCount(10),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListStocks(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListStocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("ListStocks() got = %v, want non-empty", got)
				return
			}
			if len(got) != 10 {
				t.Errorf("ListStocks() got %d stocks, want 10", len(got))
			}
		})
	}
}
