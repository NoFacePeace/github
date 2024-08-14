package sort

import "sort"

func smallestDistancePair(nums []int, k int) int {
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] > nums[j]
	})
	n := len(nums)
	bucket := make([]int, 1000001)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			dis := nums[i] - nums[j]
			bucket[dis]++
		}
	}
	for i, v := range bucket {
		k -= v
		if k <= 0 {
			return i
		}
	}
	return 0
}
