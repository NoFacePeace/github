package dp

func numberOfPermutations(n int, requirements [][]int) int {
	mod := int64(1e9) + 7
	reqMap := make(map[int]int)
	reqMap[0] = 0
	maxCnt := 0
	for _, req := range requirements {
		reqMap[req[0]] = req[1]
		if req[1] > maxCnt {
			maxCnt = req[1]
		}
	}
	if reqMap[0] != 0 {
		return 0
	}
	dp := make([][]int64, n)
	for i := range dp {
		dp[i] = make([]int64, maxCnt+1)
		for j := range dp[i] {
			dp[i][j] = -1
		}
	}
	var dfs func(int, int) int64
	dfs = func(end, cnt int) int64 {
		if cnt < 0 {
			return 0
		}
		if end == 0 {
			return 1
		}
		if dp[end][cnt] != -1 {
			return dp[end][cnt]
		}
		if r, exists := reqMap[end-1]; exists {
			if r <= cnt && cnt <= end+r {
				dp[end][cnt] = dfs(end-1, r)
				return dp[end][cnt]
			}
			return 0
		} else {
			if cnt > end {
				dp[end][cnt] = (dfs(end, cnt-1) - dfs(end-1, cnt-1-end) + dfs(end-1, cnt) + mod) % mod
			} else {
				dp[end][cnt] = (dfs(end, cnt-1) + dfs(end-1, cnt)) % mod
			}
			return dp[end][cnt]
		}
	}
	return int(dfs(n-1, reqMap[n-1]))
}
