package dp

func maximumValueSum(nums []int, k int, edges [][]int) int64 {
	n := len(nums)
	g := make([][]int, n)
	for _, e := range edges {
		u, v := e[0], e[1]
		g[u] = append(g[u], v)
		g[v] = append(g[v], u)
	}
	var dfs func(u, fa int) (int64, int64)
	dfs = func(u, fa int) (int64, int64) {
		f0, f1 := int64(0), int64(-1<<63)
		for _, v := range g[u] {
			if v != fa {
				r0, r1 := dfs(v, u)
				t := max(f1+r0, f0+r1)
				f0 = max(f0+r0, f1+r1)
				f1 = t
			}
		}
		x := int64(nums[u])
		y := int64(nums[u] ^ k)
		return max(f0+x, f1+y), max(f1+x, f0+y)
	}
	ans, _ := dfs(0, -1)
	return ans
}
