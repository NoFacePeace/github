package dp

import "sort"

func maxValue(events [][]int, k int) int {
	sort.Slice(events, func(i, j int) bool {
		return events[i][1] < events[j][1]
	})
	n := len(events)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, k+1)
	}
	for i, e := range events {
		p := sort.Search(i, func(j int) bool { return events[j][1] >= e[0] })
		for j := 1; j <= k; j++ {
			dp[i+1][j] = max(dp[i][j], dp[p][j-1]+events[i][2])
		}
	}
	return dp[n][k]
}
