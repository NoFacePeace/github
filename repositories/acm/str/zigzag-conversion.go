package str

func convert(s string, numRows int) string {
	if numRows == 1 {
		return s
	}
	arr := []byte(s)
	strs := make([][]byte, numRows)
	idx := 0
	sign := 1
	for _, v := range arr {
		strs[idx] = append(strs[idx], v)
		idx += sign
		if idx == numRows {
			idx -= 2
			sign = -1
		}
		if idx == -1 {
			idx += 2
			sign = 1
		}
	}
	ans := ""
	for _, v := range strs {
		ans += string(v)
	}
	return ans
}
