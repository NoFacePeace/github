package graph

func findJudge(n int, trust [][]int) int {
	m := len(trust)
	if m == 0 && n == 1 {
		return 1
	}
	if m == 0 {
		return -1
	}
	in := map[int]int{}
	out := map[int]int{}
	for _, line := range trust {
		v1, v2 := line[0], line[1]
		in[v2]++
		out[v1]++
	}
	for k, v := range in {
		if v != n-1 {
			continue
		}
		if out[k] != 0 {
			continue
		}
		return k
	}
	return -1
}
