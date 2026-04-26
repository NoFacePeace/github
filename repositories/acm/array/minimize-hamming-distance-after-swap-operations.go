package array

func minimumHammingDistance(source []int, target []int, allowedSwaps [][]int) int {
	n := len(source)
	m := map[int]int{}
	for i := 0; i < n; i++ {
		m[i] = i
	}
	var getP func(idx int) int
	getP = func(idx int) int {
		if m[idx] == idx {
			return idx
		}
		return getP(m[idx])
	}
	for _, allow := range allowedSwaps {
		v1, v2 := allow[0], allow[1]
		p1 := getP(v1)
		p2 := getP(v2)
		if p1 < p2 {
			m[p2] = p1
		} else {
			m[p1] = p2
		}
	}
	ans := 0
	j := 0
	for i := 0; i < n; i++ {
		if source[j] == target[i] && getP(j) == getP(i) {
			ans++
			j++
			continue
		}
		for k := j; k < n; k++ {
			if source[k] != target[i] {
				continue
			}
			if getP(k) != getP(i) {
				continue
			}
			ans++
			source[k], source[j] = source[j], source[k]
			j++
			break
		}
	}
	return n - ans
}
