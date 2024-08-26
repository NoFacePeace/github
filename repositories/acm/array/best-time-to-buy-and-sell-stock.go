package array

func maxProfit(prices []int) int {
	l := len(prices)
	if l <= 1 {
		return 0
	}
	min := prices[0]
	max := prices[0]
	ans := 0
	for _, v := range prices {
		if v > max {
			max = v
			continue
		}
		if v < min {
			val := max - min
			if val > ans {
				ans = val
			}
			min = v
			max = v
		}
	}
	val := max - min
	if val > ans {
		ans = val
	}
	return ans
}
