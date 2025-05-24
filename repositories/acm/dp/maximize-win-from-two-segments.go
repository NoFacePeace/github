package dp

import "sort"

func maximizeWin(prizePositions []int, k int) int {
	n := len(prizePositions)
	dp := make([]int, n+1)
	ans := 0
	for i := 0; i < n; i++ {
		x := sort.SearchInts(prizePositions, prizePositions[i]-k)
		ans = maxSlice(ans, i-x+1+dp[x])
		dp[i+1] = maxSlice(dp[i], i-x+1)
	}
	return ans
}
