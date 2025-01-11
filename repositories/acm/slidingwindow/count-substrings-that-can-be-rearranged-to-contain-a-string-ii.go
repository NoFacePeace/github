package slidingwindow

func validSubstringCountII(word1 string, word2 string) int64 {
	m := map[byte]int{}
	cnt := 0
	for i := 0; i < len(word2); i++ {
		m[word2[i]]--
		if m[word2[i]] == -1 {
			cnt++
		}
	}
	ans := 0
	r := 0
	for l := 0; l < len(word1); l++ {
		for r < len(word1) && cnt > 0 {
			m[word1[r]]++
			if m[word1[r]] == 0 {
				cnt--
			}
			r++
		}
		if cnt == 0 {
			ans += len(word1) - r + 1
		}
		m[word1[l]]--
		if m[word1[l]] == -1 {
			cnt++
		}
	}
	return int64(ans)
}
