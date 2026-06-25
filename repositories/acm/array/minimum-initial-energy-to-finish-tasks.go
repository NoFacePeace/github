package array

import "sort"

func minimumEffort(tasks [][]int) int {
	sort.Slice(tasks, func(a, b int) bool {
		return tasks[a][1]-tasks[a][0] < tasks[b][1]-tasks[b][0]
	})
	ans := 0
	for _, task := range tasks {
		if ans+task[0] > task[1] {
			ans = ans + task[0]
			continue
		}
		ans = task[1]
	}
	return ans
}
