package math

func minSwaps(s string) int {
	cnt, mn := 0, 0
	for _, ch := range s {
		if ch == '[' {
			cnt++
		} else {
			cnt--
			mn = min(mn, cnt)
		}
	}
	return (-mn + 1) / 2
}
