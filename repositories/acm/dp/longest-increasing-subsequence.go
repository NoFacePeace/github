package dp

func lengthOfLIS(nums []int) int {
	n := len(nums)
	dp := make([]int, n)
	mx := 0
	for i := 0; i < n; i++ {
		dp[i] = 1
		for j := 0; j < i; j++ {
			if nums[i] <= nums[j] {
				continue
			}
			dp[i] = max(dp[i], dp[j]+1)
		}
		if dp[i] > mx {
			mx = dp[i]
		}
	}
	return mx
}
