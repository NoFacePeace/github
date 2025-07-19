package dp

func maximumLengthII(nums []int, k int) int {
	dp := make([][]int, k)
	for i := range dp {
		dp[i] = make([]int, k)
	}
	res := 0
	for _, num := range nums {
		num %= k
		for pre := 0; pre < k; pre++ {
			dp[pre][num] = dp[num][pre] + 1
			res = max(res, dp[pre][num])
		}
	}
	return res
}
