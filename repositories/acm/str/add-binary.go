package str

func addBinary(a string, b string) string {
	bit := 0
	l := len(a) - 1
	r := len(b) - 1
	ans := ""
	for l >= 0 && r >= 0 {
		val := 0
		if a[l] == '1' {
			val++
		}
		if b[r] == '1' {
			val++
		}
		l--
		r--
		val += bit
		bit = 0
		if val == 0 {
			ans = "0" + ans
			continue
		}
		if val == 1 {
			ans = "1" + ans
			continue
		}
		if val == 2 {
			ans = "0" + ans
			bit = 1
			continue
		}
		bit = 1
		ans = "1" + ans
	}
	for l >= 0 {
		val := 0
		if a[l] == '1' {
			val++
		}
		val += bit
		bit = 0
		l--
		if val == 0 {
			ans = "0" + ans
			continue
		}
		if val == 1 {
			ans = "1" + ans
			continue
		}
		bit = 1
		ans = "0" + ans
	}
	for r >= 0 {
		val := 0
		if b[r] == '1' {
			val++
		}
		val += bit
		bit = 0
		r--
		if val == 0 {
			ans = "0" + ans
			continue
		}
		if val == 1 {
			ans = "1" + ans
			continue
		}
		bit = 1
		ans = "0" + ans
	}
	if bit != 0 {
		ans = "1" + ans
	}
	return ans
}
