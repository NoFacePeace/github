package hash

func remainingMethods(n int, k int, invocations [][]int) []int {
	out := map[int][]int{}
	in := map[int][]int{}
	for _, v := range invocations {
		a, b := v[0], v[1]
		if out[a] == nil {
			out[a] = []int{}
		}
		if in[b] == nil {
			in[b] = []int{}
		}
		out[a] = append(out[a], b)
		in[b] = append(in[b], a)
	}
	q := []int{k}
	visited := map[int]bool{}
	for len(q) > 0 {
		v := q[0]
		q = q[1:]
		if visited[v] {
			continue
		}
		visited[v] = true
		for _, e := range out[v] {
			if visited[e] {
				continue
			}
			q = append(q, e)
		}
	}
	if len(visited) == n {
		return []int{}
	}
	for k := range visited {
		for _, e := range in[k] {
			if !visited[e] {
				visited = map[int]bool{}
				break
			}
		}
	}

	ans := []int{}
	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		ans = append(ans, i)
	}
	return ans
}
