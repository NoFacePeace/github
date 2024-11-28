package greedy

func shortestDistanceAfterQueries(n int, queries [][]int) []int {
	roads := make([]int, n)
	for i := 0; i < n; i++ {
		roads[i] = i + 1
	}
	ans := []int{}
	dist := n - 1
	for _, query := range queries {
		u, v := query[0], query[1]
		k := roads[u]
		roads[u] = v
		for k != -1 && k < v {
			k, roads[k] = roads[k], -1
			dist--
		}
		ans = append(ans, dist)
	}
	return ans
}
