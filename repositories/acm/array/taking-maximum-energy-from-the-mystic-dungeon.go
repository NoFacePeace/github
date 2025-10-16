package array

import "math"

func maximumEnergy(energy []int, k int) int {
	n := len(energy)
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		if i-k < 0 {
			arr[i] = energy[i]
			continue
		}
		if arr[i-k] < 0 {
			arr[i] = energy[i]
			continue
		}
		arr[i] = arr[i-k] + energy[i]
	}
	mx := math.MinInt
	for i := n - k; i < n; i++ {
		mx = max(arr[i], mx)
	}
	return mx
}
