package slidingwindow

func maxProfit(prices []int, strategy []int, k int) int64 {
	ans := 0
	n := len(prices)
	for i := 0; i < n; i++ {
		ans += prices[i] * strategy[i]
	}
	sum := ans
	for i := 0; i < k; i++ {
		if i < k/2 {
			if strategy[i] == 1 {
				sum -= prices[i]
			}
			if strategy[i] == -1 {
				sum += prices[i]
			}
			continue
		}
		if strategy[i] == -1 {
			sum += prices[i] * 2
		}
		if strategy[i] == 0 {
			sum += prices[i]
		}
	}
	ans = max(ans, sum)
	for i := k; i < n; i++ {
		sum += prices[i-k] * strategy[i-k]
		sum -= prices[i-k/2]
		if strategy[i] == 0 {
			sum += prices[i]
		}
		if strategy[i] == -1 {
			sum += prices[i] * 2
		}
		ans = max(ans, sum)
	}
	return int64(ans)
}
