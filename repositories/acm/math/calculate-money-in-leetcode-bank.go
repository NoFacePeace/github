package math

func totalMoney(n int) int {
	ans := 0
	i := 1
	cnt := 0
	for cnt < n {
		limit := min(cnt+7, n)
		j := i
		for cnt < limit {
			ans += j
			j++
			cnt++
		}
		i++
	}
	return ans
}
