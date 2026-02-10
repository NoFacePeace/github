package str

func minimumDeletions(s string) int {
	a := 0
	for _, v := range s {
		if v == 'a' {
			a++
			continue
		}
	}
	ans := len(s)
	b := 0
	for _, v := range s {
		if v == 'a' {
			a--
			ans = min(ans, b+a)
			continue
		}
		ans = min(ans, b+a)
		b++
	}
	if b == 0 {
		return 0
	}
	return ans
}
