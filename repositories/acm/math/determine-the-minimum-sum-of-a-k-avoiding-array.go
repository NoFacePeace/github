package math

func minimumSum(n int, k int) int {
	arithmeticSeriesSum := func(a1, d, n int) int {
		return (a1 + a1 + (n-1)*d) * n / 2
	}
	if n <= k/2 {
		return arithmeticSeriesSum(1, 1, n)
	} else {
		return arithmeticSeriesSum(1, 1, k/2) + arithmeticSeriesSum(k, 1, n-k/2)
	}
}
