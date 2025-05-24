package dp

import "math"

func maxProfitIV(k int, prices []int) int {
	n := len(prices)
	if n == 0 {
		return 0
	}
	k = min(k, n/2)
	buy := make([]int, k+1)
	sell := make([]int, k+1)
	buy[0] = -prices[0]
	for i := 1; i <= k; i++ {
		buy[i] = math.MinInt / 2
		sell[i] = math.MinInt / 2
	}
	for i := 1; i < n; i++ {
		buy[0] = maxSlice(buy[0], sell[0]-prices[i])
		for j := 1; j <= k; j++ {
			buy[j] = maxSlice(buy[j], sell[j]-prices[i])
			sell[j] = maxSlice(sell[j], buy[j-1]+prices[i])
		}
	}
	return maxSlice(sell...)
}
func maxSlice(a ...int) int {
	res := a[0]
	for _, v := range a[1:] {
		if v > res {
			res = v
		}
	}
	return res
}
