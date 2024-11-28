package array

import (
	"math"
	"sort"
)

func minimumAverage(nums []int) float64 {
	sort.Ints(nums)
	n := len(nums)
	mn := math.MaxFloat64
	for i := 0; i < n/2; i++ {
		mn = min(mn, float64(nums[i]+nums[n-i-1])/2)
	}
	return mn
}
