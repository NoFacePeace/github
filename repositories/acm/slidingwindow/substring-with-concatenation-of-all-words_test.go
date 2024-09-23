package slidingwindow

import (
	"reflect"
	"testing"
)

func Test_findSubstring(t *testing.T) {
	type args struct {
		s     string
		words []string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		// {
		// 	name: "test",
		// 	args: args{
		// 		s:     "lingmindraboofooowingdingbarrwingmonkeypoundcake",
		// 		words: []string{"fooo", "barr", "wing", "ding", "wing"},
		// 	},
		// 	want: []int{8},
		// },
		{
			name: "test",
			args: args{
				s:     "wordgoodgoodgoodbestword",
				words: []string{"word", "good", "best", "good"},
			},
			want: []int{8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findSubstring(tt.args.s, tt.args.words); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findSubstring() = %v, want %v", got, tt.want)
			}
		})
	}
}
