package dp

func resultArray(nums []int, k int) []int64 {
	n := len(nums)
	result := make([]int64, k)
	dp := make([]int64, k)
	for i := 0; i < n; i++ {
		ndp := make([]int64, k)
		ndp[nums[i]%k]++
		for r := 0; r < k; r++ {
			ndp[(int64(r)*int64(nums[i]))%int64(k)] += dp[r]
		}
		dp = ndp
		for r := 0; r < k; r++ {
			result[r] += dp[r]
		}
	}
	return result
}
