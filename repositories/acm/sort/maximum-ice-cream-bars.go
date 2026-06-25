package sort

import "sort"

func maxIceCream(costs []int, coins int) int {
	sort.Ints(costs)
	cnt := 0
	total := 0
	n := len(costs)
	for i := 0; i < n; i++ {
		cost := costs[i]
		if total+cost > coins {
			break
		}
		total += cost
		cnt++
	}
	return cnt
}
