package array

func maximumJumps(nums []int, target int) int {
	n := len(nums)
	dp := make([]int, n)
	for i := 0; i < n; i++ {
		if i != 0 && dp[i] == 0 {
			continue
		}
		for j := i + 1; j < n; j++ {
			if nums[j]-nums[i] > target {
				continue
			}
			if nums[i]-nums[j] > target {
				continue
			}
			if dp[j] == 0 {
				dp[j] = dp[i] + 1
				continue
			}
			if dp[j] > dp[i] {
				continue
			}
			dp[j] = dp[i] + 1
		}
	}
	if dp[n-1] == 0 {
		return -1
	}
	return dp[n-1]
}
