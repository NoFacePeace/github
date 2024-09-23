package str

func longestCommonPrefix(strs []string) string {
	str := ""
	for i := 0; i < len(strs); i++ {
		if strs[i] == "" {
			return ""
		}
		if i == 0 {
			str = strs[i]
			continue
		}
		l := min(len(str), len(strs[i]))
		for j := 0; j < l; j++ {
			if str[j] != strs[i][j] {
				str = str[:j]
				break
			}
			if j == l-1 {
				str = str[:l]
			}
		}

	}
	return str
}
