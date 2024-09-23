package slidingwindow

func lengthOfLongestSubstring(s string) int {
	arr := []byte(s)
	n := len(arr)
	mx := 0
	m := map[byte]int{}
	if n == 0 {
		return 0
	}
	left, right := 0, 0
	m[s[right]]++
	for right < n {
		if right-left+1 == len(m) {
			if right-left+1 > mx {
				mx = right - left + 1
			}
			right++
			if right < n {
				m[s[right]]++
			}
		} else {
			m[s[left]]--
			if m[s[left]] == 0 {
				delete(m, s[left])
			}
			left++
		}
	}
	return mx
}
