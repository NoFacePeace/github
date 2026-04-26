package hash

import "math"

func solveQueries(nums []int, queries []int) []int {
	m := map[int]int{}
	n := len(nums)
	dist := make([]int, n)
	for i := 0; i < n; i++ {
		dist[i] = math.MaxInt
	}
	for k, v := range nums {
		if _, ok := m[v]; !ok {
			m[v] = k
			continue
		}
		idx := m[v]
		dist[k] = k - idx
		dist[idx] = min(dist[idx], k-idx)
		m[v] = k
	}
	visited := map[int]bool{}
	for k, v := range nums {
		if visited[v] {
			continue
		}
		visited[v] = true
		idx := m[v]
		if k == idx {
			dist[k] = -1
			continue
		}
		dist[k] = min(dist[k], n-idx+k)
		dist[idx] = min(dist[idx], n-idx+k)
	}
	ans := []int{}
	for _, query := range queries {
		ans = append(ans, dist[query])
	}
	return ans
}
