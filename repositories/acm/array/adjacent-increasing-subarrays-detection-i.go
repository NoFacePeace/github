package array

func hasIncreasingSubarrays(nums []int, k int) bool {
	n := len(nums)
	dp := make([]int, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			dp[i] = 1
			continue
		}
		if nums[i] > nums[i-1] {
			dp[i] = dp[i-1] + 1
		} else {
			dp[i] = 1
		}
		if dp[i] < k {
			continue
		}
		if i+1 < 2*k {
			continue
		}
		if dp[i-k] < k {
			continue
		}
		return true
	}
	return false
}
