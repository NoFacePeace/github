package greedy

import "sort"

func minSetSize(arr []int) int {
	m := map[int]int{}
	for _, v := range arr {
		m[v]++
	}
	nums := []int{}
	for _, v := range m {
		nums = append(nums, v)
	}
	sort.Ints(nums)
	n := len(nums)
	cnt := 0
	for i := n - 1; i >= 0; i-- {
		cnt += nums[i]
		if cnt >= (len(arr)+1)/2 {
			return n - i
		}
	}
	return 0
}
