package dp

func maximumLength(nums []int, k int) int {
	lenNums := len(nums)
	dp := make(map[int][]int)
	zd := make([]int, k+1)
	for i := 0; i < lenNums; i++ {
		v := nums[i]
		if _, ok := dp[v]; !ok {
			dp[v] = make([]int, k+1)
		}
		tmp := dp[v]
		for j := 0; j <= k; j++ {
			tmp[j]++
			if j > 0 {
				tmp[j] = maxSlice(tmp[j], zd[j-1]+1)
			}
		}
		for j := 0; j <= k; j++ {
			zd[j] = maxSlice(zd[j], tmp[j])
			if j > 0 {
				zd[j] = maxSlice(zd[j], zd[j-1])
			}
		}
	}
	return zd[k]
}
