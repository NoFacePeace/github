package sort

import "sort"

func countDays(days int, meetings [][]int) int {
	sort.Slice(meetings, func(a, b int) bool {
		if meetings[a][0] == meetings[b][0] {
			return meetings[a][1] < meetings[b][1]
		}
		return meetings[a][0] < meetings[b][0]
	})
	n := len(meetings)
	ans := 0
	end := 0
	for i := 0; i < n; i++ {
		if i == 0 {
			ans += meetings[i][0] - 1
			end = meetings[i][1]
		} else if end < meetings[i][0] {
			ans += meetings[i][0] - end - 1
			end = meetings[i][1]
		}
		end = max(end, meetings[i][1])
		if i == n-1 {
			ans += days - end
		}
	}
	return ans
}
