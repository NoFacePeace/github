package search

func countCompleteComponents(n int, edges [][]int) int {
	adj := make([][]int, n)
	for i := range adj {
		adj[i] = []int{}
	}
	for _, edg := range edges {
		u, v := edg[0], edg[1]
		adj[u] = append(adj[u], v)
		adj[v] = append(adj[v], u)
	}
	visited := map[int]bool{}
	ans := 0
	cnt := 0
	var dfs func(idx, l int)
	dfs = func(idx, l int) {
		if len(adj[idx]) == l {
			cnt++
		}
		for _, v := range adj[idx] {
			if visited[v] {
				continue
			}
			visited[v] = true
			dfs(v, l)
		}
	}
	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		cnt = 0
		pre := len(visited)
		visited[i] = true
		dfs(i, len(adj[i]))
		cur := len(visited)
		if cur-pre == len(adj[i])+1 && cnt == cur-pre {
			ans++
		}
	}
	return ans
}
