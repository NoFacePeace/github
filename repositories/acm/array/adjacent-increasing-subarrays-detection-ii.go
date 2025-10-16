package array

func maxIncreasingSubarrays(nums []int) int {
	n := len(nums)
	dp := make([]int, n)
	ans := 0
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
		for j := ans + 1; j <= dp[i]; j++ {
			if i-j < 0 {
				break
			}
			if dp[i-j] < j {
				break
			}
			ans = j
		}
	}
	return ans
}
