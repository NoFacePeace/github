package array

func maximumDifference(nums []int) int {
	n := len(nums)
	if n < 2 {
		return -1
	}
	mn := nums[0]
	ans := -1
	for i := 1; i < n; i++ {
		dist := nums[i] - mn
		if dist > 0 {
			ans = max(ans, dist)
		}
		mn = min(mn, nums[i])
	}
	return ans
}
