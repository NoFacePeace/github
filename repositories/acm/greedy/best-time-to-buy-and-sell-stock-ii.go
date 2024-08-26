package greedy

func maxProfit(prices []int) int {
	l := len(prices)
	if l <= 1 {
		return 0
	}
	sum := 0
	min := prices[0]
	for _, v := range prices {
		if v > min {
			sum += v - min
		}
		min = v
	}
	return sum
}
