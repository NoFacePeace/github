package pointer

import "sort"

func triangleNumber(nums []int) int {
	n := len(nums)
	sort.Ints(nums)
	ans := 0
	for i, v := range nums {
		k := i
		for j := i + 1; j < n; j++ {
			for k+1 < n && nums[k+1] < v+nums[j] {
				k++
			}
			ans += max(k-j, 0)
		}
	}
	return ans
}
