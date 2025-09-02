package bitwise

import (
	"reflect"
	"testing"
)

func Test_productQueries(t *testing.T) {
	type args struct {
		n       int
		queries [][]int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			args: args{
				n:       15,
				queries: [][]int{{0, 1}, {2, 2}, {0, 3}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := productQueries(tt.args.n, tt.args.queries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
