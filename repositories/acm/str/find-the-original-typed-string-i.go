package str

func possibleStringCount(word string) int {
	ans := 1
	var pre rune
	cnt := 0
	for _, v := range word {
		if v != pre {
			ans += cnt
			pre = v
			cnt = 0
			continue
		}
		cnt++
	}
	ans += cnt
	return ans
}
