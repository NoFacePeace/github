package set

func processQueries(c int, connections [][]int, queries [][]int) []int {
	m := map[int]int{}
	m1 := map[int]int{}
	offline := map[int]bool{}
	merge := func(a, b int) {
		for m[a] != 0 {
			a = m[a]
		}
		for m[b] != 0 {
			b = m[b]
		}
		a, b = min(a, b), max(a, b)
		m[b] = a
	}
	merge1 := func(a, b int) {
		for m1[a] != 0 {
			a = m1[a]
		}
		for m1[b] != 0 {
			b = m1[b]
		}
		a, b = min(a, b), max(a, b)
		m1[a] = b
	}
	search := func(x int) int {
		ret := -1
		for m1[x] != 0 {
			x = m1[x]
		}
		for m[x] != 0 {
			x = m[x]
			if !offline[x] {
				ret = x
			}
		}
		return ret
	}
	for _, connection := range connections {
		n1, n2 := connection[0], connection[1]
		merge(n1, n2)
		merge1(n1, n2)
	}
	ans := []int{}
	for _, query := range queries {
		op, x := query[0], query[1]
		if op == 2 {
			offline[x] = true
			continue
		}
		if !offline[x] {
			ans = append(ans, x)
			continue
		}
		x = search(x)
		ans = append(ans, x)
	}
	return ans
}
