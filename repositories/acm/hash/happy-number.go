package hash

func isHappy(n int) bool {
	m := map[int]bool{}
	for !m[n] {
		if n == 1 {
			return true
		}
		m[n] = true
		sum := 0
		for n != 0 {
			mod := n % 10
			n /= 10
			sum += mod * mod
		}
		n = sum
	}
	return false
}
