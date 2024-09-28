package str

func maximumSubsequenceCount(text string, pattern string) int64 {
	n := len(text)
	first := make([]int, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			if text[i] == pattern[0] {
				first[i] = 1
			}
			continue
		}
		if text[i] == pattern[0] {
			first[i] += first[i-1] + 1
		} else {
			first[i] += first[i-1]
		}
	}
	cnt := 0
	two := 0
	for i := n - 1; i >= 0; i-- {
		if i != 0 && text[i] == pattern[1] {
			cnt += first[i-1]
		}
		if text[i] == pattern[1] {
			two++
		}
	}
	return int64(cnt + max(two, first[n-1]))
}
