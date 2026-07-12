package sort

import "sort"

func arrayRankTransform(arr []int) []int {
	n := len(arr)
	nums := [][2]int{}
	for k, v := range arr {
		nums = append(nums, [2]int{k, v})
	}
	sort.Slice(nums, func(a, b int) bool {
		return nums[a][1] < nums[b][1]
	})
	ans := make([]int, n)
	cnt := 1
	last := 0
	for idx, num := range nums {
		k, v := num[0], num[1]
		if idx == 0 {
			last = v
			ans[k] = cnt
			continue
		}
		if v == last {
			ans[k] = cnt
			continue
		}
		cnt++
		ans[k] = cnt
		last = v
	}
	return ans
}
