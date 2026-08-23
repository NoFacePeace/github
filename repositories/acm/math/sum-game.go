package math

func sumGame(num string) bool {
	n := len(num)
	left := 0
	right := 0
	cnt := 0
	for i := 0; i < n; i++ {
		c := num[i]
		if c == '?' {
			if i < n/2 {
				cnt++
			} else {
				cnt--
			}
			continue
		}
		val := int(c - '0')
		if i < n/2 {
			left += val
		} else {
			right += val
		}
	}
	if cnt < 0 {
		left, right = right, left
		cnt = -cnt
	}
	if left <= right {
		if (cnt+1)/2*9+left > right {
			return true
		}
		if cnt/2*9+left < right {
			return true
		}
		return false
	}
	return true
}
