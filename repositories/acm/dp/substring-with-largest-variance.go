package dp

func largestVariance(s string) int {
	pos := make(map[rune][]int)
	for i, ch := range s {
		pos[ch] = append(pos[ch], i)
	}
	ans := 0
	for c0, pos0 := range pos {
		for c1, pos1 := range pos {
			if c0 != c1 {
				i, j := 0, 0
				f, g := 0, -1<<63
				for i < len(pos0) || j < len(pos1) {
					if j == len(pos1) || (i < len(pos0) && pos0[i] < pos1[j]) {
						f, g = maxSlice(f, 0)+1, g+1
						i++
					} else {
						f, g = maxSlice(f, 0)-1, maxSlice(f, g, 0)-1
						j++
					}
					ans = maxSlice(ans, g)
				}
			}
		}
	}
	return ans
}
