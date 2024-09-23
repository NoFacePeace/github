package hash

func isAnagram(s string, t string) bool {
	sb := []byte(s)
	sm := map[byte]int{}
	for _, v := range sb {
		sm[v]++
	}

	tb := []byte(t)
	tm := map[byte]int{}
	for _, v := range tb {
		tm[v]++
	}
	if len(sm) != len(tm) {
		return false
	}
	for k, v := range sm {
		if v != tm[k] {
			return false
		}
	}
	return true
}
