package bitwise

func reverseBits(n int) int {
	num := 0
	cnt := 0
	for n != 0 {
		bit := n & 1
		num = num << 1
		num += bit
		n = n >> 1
		cnt++
	}
	if cnt != 32 {
		num = num << (32 - cnt)
	}
	return num
}
