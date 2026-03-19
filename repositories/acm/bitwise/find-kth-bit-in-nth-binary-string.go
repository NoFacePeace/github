package bitwise

func findKthBit(n int, k int) byte {
	l := 0
	for i := 0; i < n; i++ {
		l = l*2 + 1
	}
	cnt := 0
	for n > 0 {
		mid := l / 2
		if k == mid {
			return '1'
		}
		if k > mid {
			k = k - mid
			cnt++
		}
		l = l - mid
		n--
	}
	if cnt%2 == 0 {
		return '0'
	}
	return '1'
}
