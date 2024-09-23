package array

import "testing"

func Test_latestTimeCatchTheBus(t *testing.T) {
	type args struct {
		buses      []int
		passengers []int
		capacity   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				buses:      []int{3},
				passengers: []int{2, 3},
				capacity:   2,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := latestTimeCatchTheBus(tt.args.buses, tt.args.passengers, tt.args.capacity); got != tt.want {
				t.Errorf("latestTimeCatchTheBus() = %v, want %v", got, tt.want)
			}
		})
	}
}
