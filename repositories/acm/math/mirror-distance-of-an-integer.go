package math

func mirrorDistance(n int) int {
	mirror := 0
	tmp := n
	for tmp != 0 {
		bit := tmp % 10
		tmp /= 10
		mirror = mirror*10 + bit
	}
	abs := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	return abs(mirror, n)
}
