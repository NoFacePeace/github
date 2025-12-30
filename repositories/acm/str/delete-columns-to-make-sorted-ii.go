package str

func minDeletionSizeII(strs []string) int {
	n := len(strs)
	if n == 0 {
		return 0
	}
	m := len(strs[0])
	if m == 0 {
		return 0
	}
	del := map[int]bool{}
	ok := map[int]bool{}
	var f func(idx int)
	f = func(idx int) {
		if idx == m {
			return
		}
		tmp := map[int]bool{}
		for i := 0; i < n-1; i++ {
			if ok[i] {
				continue
			}
			if strs[i][idx] > strs[i+1][idx] {
				del[idx] = true
				f(idx + 1)
				return
			}
			if strs[i][idx] < strs[i+1][idx] {
				tmp[i] = true
			}
		}
		for k, v := range tmp {
			ok[k] = v
		}
		if len(ok) == n {
			return
		}
		f(idx + 1)
	}
	f(0)
	return len(del)
}
