package math

func flowerGame(n int, m int) int64 {
	sum := n * m
	sum -= (n / 2) * (m / 2)
	sum -= (n/2 + n%2) * (m/2 + m%2)
	return int64(sum)
}
