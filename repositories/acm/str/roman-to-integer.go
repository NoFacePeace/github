package str

func romanToInt(s string) int {
	m := map[string]int{
		"I":  1,
		"V":  5,
		"X":  10,
		"L":  50,
		"C":  100,
		"D":  500,
		"M":  1000,
		"IV": 4,
		"IX": 9,
		"XL": 40,
		"XC": 90,
		"CD": 400,
		"CM": 900,
	}
	n := len(s) - 1
	ans := 0
	for n >= 0 {
		if n-1 >= 0 {
			if val, ok := m[s[n-1:n+1]]; ok {
				ans += val
				n -= 2
				continue
			}
		}
		val := m[s[n:n+1]]
		ans += val
		n--
	}
	return ans
}
