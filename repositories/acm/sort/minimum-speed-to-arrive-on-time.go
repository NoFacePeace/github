package sort

import "math"

func minSpeedOnTime(dist []int, hour float64) int {
	n := len(dist)
	hr := int(math.Round(hour * 100))
	if hr <= (n-1)*100 {
		return -1
	}
	l, r := 1, 10000000
	for l < r {
		mid := l + (r-l)/2
		t := 0
		for i := 0; i < n-1; i++ {
			t += (dist[i]-1)/mid + 1
		}
		t *= mid
		t += dist[n-1]
		if t*100 <= hr*mid {
			r = mid
		} else {
			l = mid + 1
		}
	}
	return l
}
