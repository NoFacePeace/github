package array

import "sort"

func threeSum(nums []int) [][]int {
	n := len(nums)
	ans := [][]int{}
	sort.Ints(nums)
	for i := 0; i < n; i++ {
		if i != 0 && nums[i] == nums[i-1] {
			continue
		}
		k := n - 1
		for j := i + 1; j < k; j++ {
			if j != i+1 && nums[j] == nums[j-1] {
				continue
			}
			if nums[i]+nums[j]+nums[k] > 0 {
				k--
			}
			if j != k && nums[i]+nums[j]+nums[k] == 0 {
				ans = append(ans, []int{nums[i], nums[j], nums[k]})
				continue
			}

		}
	}
	return ans
}
