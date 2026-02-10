package sort

import (
	"math"
	"sort"
)

func minimumDifference(nums []int, k int) int {
	sort.Ints(nums)
	n := len(nums)
	ans := math.MaxInt
	for i := 0; i <= n-k; i++ {
		ans = min(ans, nums[i+k-1]-nums[i])
	}
	return ans
}
