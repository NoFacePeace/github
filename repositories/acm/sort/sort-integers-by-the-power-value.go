package sort

import "sort"

func getKth(lo int, hi int, k int) int {
	m := map[int]int{}
	var compute func(num int) int
	compute = func(num int) int {
		if m[num] != 0 {
			return m[num]
		}
		if num == 1 {
			return 0
		}
		if num%2 == 0 {
			m[num] = compute(num/2) + 1
		} else {
			m[num] = compute(3*num+1) + 1
		}
		return m[num]
	}
	nums := []int{}
	for i := lo; i <= hi; i++ {
		nums = append(nums, i)
	}
	sort.Slice(nums, func(i, j int) bool {
		if compute(nums[i]) < compute(nums[j]) {
			return true
		}
		if compute(nums[i]) == compute(nums[j]) {
			return nums[i] < nums[j]
		}
		return false
	})
	return nums[k-1]
}
