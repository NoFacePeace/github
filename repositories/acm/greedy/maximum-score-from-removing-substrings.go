package greedy

func maximumGain(s string, x int, y int) int {
	a, b := 'a', 'b'
	if x < y {
		x, y = y, x
		a, b = b, a
	}
	cnt1 := 0
	cnt2 := 0
	ans := 0
	for _, c := range s {
		if c == a {
			cnt1++
		} else if c == b {
			if cnt1 > 0 {
				ans += x
				cnt1--
			} else {
				cnt2++
			}
		} else {
			ans += min(cnt1, cnt2) * y
			cnt1 = 0
			cnt2 = 0
		}
	}
	ans += min(cnt1, cnt2) * y
	return ans
}
