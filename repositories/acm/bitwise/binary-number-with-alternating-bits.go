package bitwise

func hasAlternatingBits(n int) bool {
	last := -1
	for n != 0 {
		bit := n & 1
		if bit == last {
			return false
		}
		last = bit
		n = n >> 1
	}
	return true
}
