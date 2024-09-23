package str

func lengthOfLastWord(s string) int {
	cnt := 0
	pre := ' '
	for _, v := range s {
		if v == ' ' {
			pre = ' '
			continue
		}
		if pre == ' ' {
			cnt = 1
			pre = v
			continue
		}
		cnt++
	}
	return cnt
}
