package sort

import "sort"

func removeCoveredIntervals(intervals [][]int) int {
	sort.Slice(intervals, func(a, b int) bool {
		if intervals[a][0] == intervals[b][0] {
			return intervals[a][1] > intervals[b][1]
		}
		return intervals[a][0] < intervals[b][0]
	})
	n := len(intervals)
	ans := 0
	last := 0
	for i := 0; i < n; i++ {
		if i == 0 {
			ans++
			last = i
			continue
		}
		interval := intervals[i]
		if interval[1] <= intervals[last][1] {
			continue
		}
		last = i
		ans++
	}
	return ans
}
