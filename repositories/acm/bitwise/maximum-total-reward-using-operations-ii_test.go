package bitwise

import "testing"

func Test_maxTotalReward(t *testing.T) {
	type args struct {
		rewardValues []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				rewardValues: []int{1, 1, 3, 3},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxTotalReward(tt.args.rewardValues); got != tt.want {
				t.Errorf("maxTotalReward() = %v, want %v", got, tt.want)
			}
		})
	}
}
