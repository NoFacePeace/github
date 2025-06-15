package array

func maxAdjacentDistance(nums []int) int {
	abs := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	n := len(nums)
	if n < 2 {
		return 0
	}
	ans := abs(nums[0], nums[n-1])
	for i := 0; i < n-1; i++ {
		ans = max(ans, abs(nums[i], nums[i+1]))
	}
	return ans
}
