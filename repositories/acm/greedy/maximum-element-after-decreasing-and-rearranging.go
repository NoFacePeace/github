package greedy

import "sort"

func maximumElementAfterDecrementingAndRearranging(arr []int) int {
	sort.Ints(arr)
	n := len(arr)
	inc := 0
	for i := 0; i < n; i++ {
		num := arr[i]
		if num == inc {
			continue
		}
		inc++
	}
	return inc
}
