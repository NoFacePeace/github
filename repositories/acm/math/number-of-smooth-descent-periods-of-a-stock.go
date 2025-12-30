package math

func getDescentPeriods(prices []int) int64 {
	n := len(prices)
	if n == 0 {
		return 0
	}
	cnt := 1
	ans := 0
	for i := 1; i < n; i++ {
		if prices[i]+1 == prices[i-1] {
			cnt++
			continue
		}
		ans += cnt * (cnt + 1) / 2
		cnt = 1
	}
	ans += cnt * (cnt + 1) / 2
	return int64(ans)
}
