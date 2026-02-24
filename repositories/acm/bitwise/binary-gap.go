package bitwise

func binaryGap(n int) int {
	ans := 0
	cnt := 0
	last := -1
	for n != 0 {
		bit := n & 1
		n = n >> 1
		cnt++
		if bit == 0 {
			continue
		}
		if last == -1 {
			last = cnt
			continue
		}
		ans = max(ans, cnt-last)
		last = cnt
	}
	return ans
}
