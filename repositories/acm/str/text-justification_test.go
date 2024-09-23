package str

import (
	"reflect"
	"testing"
)

func Test_split(t *testing.T) {
	type args struct {
		words    []string
		maxWidth int
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "test",
			args: args{
				words:    []string{"This", "is", "an", "example", "of", "text", "justification."},
				maxWidth: 16,
			},
			want: [][]string{
				{"This", "is", "an"},
				{"example", "of", "text"},
				{"justification."},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := split(tt.args.words, tt.args.maxWidth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_join(t *testing.T) {
	type args struct {
		words    []string
		maxWidth int
		end      bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				words:    []string{"This", "is", "an"},
				maxWidth: 16,
				end:      false,
			},
			want: "This    is    an",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := join(tt.args.words, tt.args.maxWidth, tt.args.end); got != tt.want {
				t.Errorf("join() = %v, want %v", got, tt.want)
			}
		})
	}
}
