package sort

import "sort"

func maxTwoEvents(events [][]int) int {
	n := len(events)
	if n == 0 {
		return 0
	}
	times := []int{}
	vals := []int{}
	sort.Slice(events, func(a, b int) bool {
		return events[a][1] < events[b][1]
	})
	ans := 0
	for i := 0; i < n; i++ {
		event := events[i]
		start, end, val := event[0], event[1], event[2]
		if i == 0 {
			times = append(times, end)
			vals = append(vals, val)
			ans = max(ans, val)
			continue
		}
		idx := sort.Search(len(vals), func(i int) bool {
			return times[i] >= start
		})
		if idx == 0 {
			ans = max(ans, val)
		} else {
			ans = max(ans, val+vals[idx-1])
		}
		times = append(times, end)
		vals = append(vals, max(val, vals[len(vals)-1]))
	}
	return ans
}
