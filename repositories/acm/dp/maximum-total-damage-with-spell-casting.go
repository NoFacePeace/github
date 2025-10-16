package dp

import "sort"

func maximumTotalDamage(power []int) int64 {
	n := len(power)
	dp := make([]int, n)
	mx := make([]int, n)
	sort.Ints(power)
	for i := 0; i < n; i++ {
		if i == 0 {
			dp[i] = power[i]
			mx[i] = dp[i]
			continue
		}
		if power[i-1] == power[i] {
			dp[i] = dp[i-1] + power[i]
			mx[i] = max(mx[i-1], dp[i])
			continue
		}
		idx := sort.Search(i, func(idx int) bool {
			return power[idx] >= power[i]-2
		})
		if idx == 0 {
			dp[i] = power[i]
			mx[i] = max(mx[i-1], dp[i])
			continue
		}
		if idx == i {
			dp[i] = mx[i-1] + power[i]
			mx[i] = dp[i]
			continue
		}
		dp[i] = mx[idx-1] + power[i]
		mx[i] = max(mx[i-1], dp[i])
	}

	return int64(mx[n-1])
}
