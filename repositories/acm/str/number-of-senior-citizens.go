package str

func countSeniors(details []string) int {
	cnt := 0
	for _, v := range details {
		if v[11] > '6' {
			cnt++
			continue
		}
		if v[11] == '6' && v[12] != '0' {
			cnt++
			continue
		}
	}
	return cnt
}
