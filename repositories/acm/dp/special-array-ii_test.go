package dp

import (
	"reflect"
	"testing"
)

func Test_isArraySpecial(t *testing.T) {
	type args struct {
		nums    []int
		queries [][]int
	}
	tests := []struct {
		name string
		args args
		want []bool
	}{
		{
			name: "test",
			args: args{
				nums:    []int{1, 8},
				queries: [][]int{{1, 1}},
			},
			want: []bool{true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isArraySpecial(tt.args.nums, tt.args.queries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isArraySpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}
