package array

import "sort"

func medianOfUniquenessArray(nums []int) int {
	n := len(nums)
	arr := []int{}
	for i := 1; i <= n; i++ {
		m := map[int]int{}
		j := 0
		for j < i {
			m[nums[j]]++
			j++
		}
		arr = append(arr, len(m))
		for j < n {
			m[nums[j]]++
			m[nums[j-i]]--
			if m[nums[j-i]] == 0 {
				delete(m, nums[j-i])
			}
			arr = append(arr, len(m))
			j++
		}
	}
	sort.Ints(arr)
	if n%2 == 0 {
		return arr[n/2-1]
	}
	return arr[n/2]
}
