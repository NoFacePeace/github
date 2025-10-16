package array

import "sort"

func largestPerimeter(nums []int) int {
	sort.Ints(nums)
	n := len(nums)
	ans := 0
	j := 0
	for i := 0; i < n; i++ {
		j = max(j, i+1)
		for k := i + 2; k < n; k++ {
			if k <= j {
				continue
			}
			for j < k {
				if nums[i]+nums[j] > nums[k] && nums[k]-nums[j] < nums[i] {
					ans = max(ans, nums[i]+nums[j]+nums[k])
					j++
					if j == k {
						j--
						break
					}
					continue
				}
				break
			}
		}
	}
	return ans
}
