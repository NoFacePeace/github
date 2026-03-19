package bitwise

func bitwiseComplement(n int) int {
	if n == 0 {
		return 1
	}
	num := 0
	cnt := 0
	for n > 0 {
		bit := n & 1
		if bit == 0 {
			bit = 1 << cnt
			num += bit
		}
		n = n >> 1
		cnt++
	}
	return num
}
