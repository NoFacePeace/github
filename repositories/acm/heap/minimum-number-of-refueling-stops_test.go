package heap

import "testing"

func Test_minRefuelStops(t *testing.T) {
	type args struct {
		target    int
		startFuel int
		stations  [][]int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				target:    100,
				startFuel: 10,
				stations: [][]int{
					{10, 60},
					{20, 30},
					{30, 30},
					{60, 40},
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minRefuelStops(tt.args.target, tt.args.startFuel, tt.args.stations); got != tt.want {
				t.Errorf("minRefuelStops() = %v, want %v", got, tt.want)
			}
		})
	}
}
