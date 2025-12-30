package sort

import "sort"

func minimumBoxes(apple []int, capacity []int) int {
	sum := 0
	for _, v := range apple {
		sum += v
	}
	sort.Ints(capacity)
	n := len(capacity)
	ans := 0
	for i := n - 1; i >= 0; i-- {
		if sum <= 0 {
			break
		}
		sum -= capacity[i]
		ans++
	}
	return ans
}
