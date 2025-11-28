package array

import "sort"

func maxFrequency(nums []int, k int, numOperations int) int {
	sort.Ints(nums)
	n := len(nums)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	mn := nums[0]
	mx := nums[n-1]
	ans := 1
	m := map[int]int{}
	for _, v := range nums {
		m[v]++
	}
	for i := mn; i <= mx; i++ {
		left := sort.Search(n, func(j int) bool {
			return nums[j] >= i-k
		})
		right := sort.Search(n, func(j int) bool {
			return nums[j] > i+k
		})
		cnt := right - left - m[i]
		cnt = min(numOperations, cnt)
		ans = max(ans, cnt+m[i])
	}
	return ans
}
