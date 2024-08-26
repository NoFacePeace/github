package str

func checkRecord(s string) bool {
	a := 0
	l := 0
	for _, v := range s {
		if v == 'A' {
			a++
			l = 0
		}
		if v == 'L' {
			l++
		}
		if v == 'P' {
			l = 0
		}
		if a >= 2 {
			return false
		}
		if l >= 3 {
			return false
		}
	}
	return true
}
