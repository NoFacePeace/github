package slidingwindow

func findSubstring(s string, words []string) []int {
	m := len(words)
	if m == 0 {
		return []int{}
	}
	n := len(words[0])

	stats := map[string]int{}
	for _, v := range words {
		stats[v]++
	}

	l, r := 0, 0
	exist := map[string]int{}
	cnt := 0
	ans := []int{}
	for r+n <= len(s) {
		word := s[r : r+n]
		if _, ok := stats[word]; !ok {
			l++
			r = l
			exist = map[string]int{}
			cnt = 0
			continue
		}
		if exist[word]+1 > stats[word] {
			l++
			r = l
			exist = map[string]int{}
			cnt = 0
			continue
		}
		exist[word]++
		cnt++
		if cnt == m {
			ans = append(ans, l)
			l++
			r = l
			exist = map[string]int{}
			cnt = 0
			continue
		}
		r += n
	}
	return ans
}
