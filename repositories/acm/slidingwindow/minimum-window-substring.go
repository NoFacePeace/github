package slidingwindow

func minWindow(s string, t string) string {
	tb := []byte(t)
	tm := map[byte]int{}
	for _, v := range tb {
		tm[v]++
	}

	sb := []byte(s)
	sm := map[byte]int{}

	l, r := 0, 0
	ans := ""
	cnt := 0
	for r < len(sb) {
		c := sb[r]
		if _, ok := tm[c]; !ok {
			r++
			continue
		}
		sm[c]++
		if sm[c] > tm[c] {
			r++
			continue
		}
		cnt++
		if cnt != len(tb) {
			r++
			continue
		}
		tmp := s[l : r+1]
		if ans == "" || len(ans) > len(tmp) {
			ans = tmp
		}
		if l == r {
			sm = map[byte]int{}
			l++
			r++
			cnt = 0
			continue
		}
		sm[c]--
		if sm[c] == 0 {
			delete(sm, c)
		}
		cnt--
		c = sb[l]
		l++
		if _, ok := tm[c]; !ok {
			continue
		}
		sm[c]--
		if sm[c] < tm[c] {
			cnt--
		}
		if sm[c] == 0 {
			delete(sm, c)
		}
	}
	return ans
}
