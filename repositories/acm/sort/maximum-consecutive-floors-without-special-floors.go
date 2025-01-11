package sort

import "sort"

func maxConsecutive(bottom int, top int, special []int) int {
	sort.Ints(special)
	start := bottom
	ans := 0
	for _, v := range special {
		ans = max(ans, v-start)
		start = v + 1
	}
	ans = max(ans, top-start+1)
	return ans
}
