package array

import "testing"

func Test_medianOfUniquenessArray(t *testing.T) {
	type args struct {
		nums []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				nums: []int{3, 4, 3, 4, 5},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := medianOfUniquenessArray(tt.args.nums); got != tt.want {
				t.Errorf("medianOfUniquenessArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
