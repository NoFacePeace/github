package dp

func maxProfit(prices []int) int {
	buy1, sell1 := -prices[0], 0
	buy2, sell2 := -prices[0], 0
	for i := 0; i < len(prices); i++ {
		buy1 = maxSlice(buy1, -prices[i])
		sell1 = maxSlice(sell1, buy1+prices[i])
		buy2 = maxSlice(buy2, sell1-prices[i])
		sell2 = maxSlice(sell2, buy2+prices[i])
	}
	return sell2
}
