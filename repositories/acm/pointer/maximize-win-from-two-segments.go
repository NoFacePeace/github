package pointer

func maximizeWin(prizePositions []int, k int) int {
	n := len(prizePositions)
	dp := make([]int, n+1)
	ans := 0
	for left, right := 0, 0; right < n; right++ {
		for prizePositions[right]-prizePositions[left] > k {
			left++
		}
		ans = max(ans, right-left+1+dp[left])
		dp[right+1] = max(dp[right], right-left+1)
	}
	return ans
}
