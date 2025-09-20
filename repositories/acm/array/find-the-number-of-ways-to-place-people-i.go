package array

import (
	"math"
	"sort"
)

func numberOfPairs(points [][]int) int {
	sort.Slice(points, func(a, b int) bool {
		if points[a][0] == points[b][0] {
			return points[a][1] < points[b][1]
		}
		return points[a][0] > points[b][0]
	})
	n := len(points)
	ans := 0
	for i := 0; i < n; i++ {
		mn := math.MaxInt
		for j := i + 1; j < n; j++ {
			if points[j][1] < points[i][1] {
				continue
			}
			if points[j][1] < mn {
				ans++
				mn = points[j][1]
			}
		}
	}
	return ans
}
