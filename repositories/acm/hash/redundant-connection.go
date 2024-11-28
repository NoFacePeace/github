package hash

func findRedundantConnection(edges [][]int) []int {
	m := map[int]bool{}
	ans := make([]int, 2)
	for _, edge := range edges {
		v1, v2 := edge[0], edge[1]
		if m[v1] && m[v2] {
			ans[0] = v1
			ans[1] = v2
		}
		m[v1] = true
		m[v2] = true
	}
	return ans
}
