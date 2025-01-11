package backtrack

import "math"

func minimumCost(m int, n int, horizontalCut []int, verticalCut []int) int {
	if m == 1 && n == 1 {
		return 0
	}
	mx := math.MaxInt
	for i := 1; i < n; i++ {
		mx = min(mx, minimumCost(m, n-i, horizontalCut, verticalCut[i:])+minimumCost(m, i, horizontalCut, verticalCut[:i-1])+verticalCut[i-1])
	}
	for i := 1; i < m; i++ {
		mx = min(mx, minimumCost(m-i, n, horizontalCut[i:], verticalCut)+minimumCost(i, n, horizontalCut[:i-1], verticalCut)+horizontalCut[i-1])
	}

	return mx
}
