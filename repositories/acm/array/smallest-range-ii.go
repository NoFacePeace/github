package array

import (
	"sort"
)

func smallestRangeII(nums []int, k int) int {
	sort.Ints(nums)
	n := len(nums)
	mi, ma := nums[0], nums[n-1]
	ans := ma - mi
	for i := 0; i < n-1; i++ {
		a, b := nums[i], nums[i+1]
		ans = min(ans, max(ma-k, a+k)-min(mi+k, b-k))
	}
	return ans
}
